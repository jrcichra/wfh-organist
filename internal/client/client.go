package client

import (
	"bufio"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/jrcichra/wfh-organist/internal/common"
	"github.com/jrcichra/wfh-organist/internal/parser/channels"
	"github.com/jrcichra/wfh-organist/internal/player"
	"github.com/jrcichra/wfh-organist/internal/serial"
	"github.com/jrcichra/wfh-organist/internal/types"
	"github.com/jrcichra/wfh-organist/internal/volume"
	"github.com/jrcichra/wfh-organist/pkg/timer"
	"gitlab.com/gomidi/midi"
	"gitlab.com/gomidi/midi/midimessage/channel"
	"gitlab.com/gomidi/midi/reader"
	"gitlab.com/gomidi/midi/writer"
	driver "gitlab.com/gomidi/rtmididrv"
)

func dial(serverIP string, serverPort int, protocol string) net.Conn {
	// keep trying to connect
	for {
		serverStr := fmt.Sprintf("%s:%d", serverIP, serverPort)
		log.Println("Connecting to " + serverStr + "...")
		conn, err := net.Dial(protocol, serverStr)
		if err != nil {
			log.Println(err)
			time.Sleep(time.Second * 5)
			continue
		}
		log.Println("Connected to", serverStr)
		return conn
	}
}

func Client(midiPort int, serverIP string, serverPort int, protocol string, stdinMode bool, delay int, file string, midiTuxChan chan types.MidiTuxMessage, profile string, dontControlVolume bool) {

	// read the csv
	csvRecords := channels.ReadFile(profile + "channels.csv")

	notesChan := make(chan interface{})
	stopChan := make(chan bool)

	drv, err := driver.New()
	common.Must(err)
	// make sure to close all open ports at the end
	defer drv.Close()

	in := common.GetMidiInput(drv, midiPort)
	out := common.GetMidiOutput(drv, midiPort)

	common.SetupCloseHandler(out, stopChan)

	// make a writer for each channel
	writers := make([]*writer.Writer, 16)
	var i uint8
	for ; i < 16; i++ {
		writers[i] = writer.New(out)
		writers[i].SetChannel(i)
	}

	if !dontControlVolume {
		volume.SetVolume(common.HIGH_VOLUME)
	}

	// http server for debug pprof
	go http.ListenAndServe(":8081", nil)

	// in either mode read the serial for now
	go serial.ReadSerial(notesChan)

	if stdinMode {
		go stdinClient(notesChan)
	}

	if file == "" {
		go midiClient(midiPort, delay, notesChan, in)
	} else {
		go player.PlayMidiFile(notesChan, file, stopChan, false)
	}

	// things that would need a new connection if the connection was lost
	for {
		wg := &sync.WaitGroup{}
		conn := dial(serverIP, serverPort, protocol)
		closedChan := make(chan struct{})
		// ability to send notes
		wg.Add(1)
		go sendNotesClient(wg, closedChan, conn, delay, notesChan, csvRecords, dontControlVolume)
		// ability to get your own notes back
		wg.Add(1)
		go midiClientFeedback(wg, closedChan, conn, writers, out, midiTuxChan)
		wg.Wait() // reconnect and respawn things on the waitgroup coming back
	}
}

func stdinClient(notesChan chan interface{}) {

	channel := make(chan types.Raw)

	//get stdin in a goroutine
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		// word at a time
		scanner.Split(bufio.ScanWords)
		// keep count, send 3 at a time
		count := 0
		// hold bytes
		bytes := make([]byte, 0)
		for scanner.Scan() {
			//convert token string to hex code
			text := scanner.Text()
			// each token must be size 2
			if len(text) != 2 {
				log.Println("Token must be size 2")
				bytes = make([]byte, 0)
				count = 0
				continue
			}
			hexToken, err := hex.DecodeString(text)
			common.Must(err)
			// append to bytes
			bytes = append(bytes, hexToken...)
			if count >= 2 {
				//send hex code to channel
				channel <- types.Raw{
					Time: time.Now(),
					Data: bytes,
				}
				count = 0
				// clear bytes
				bytes = make([]byte, 0)
			} else {
				count++
			}
		}
		if err := scanner.Err(); err != nil {
			log.Println(err)
		}
		// when the scanner is done, quit the program
		os.Exit(0)
	}()

	// read from the channel and send to server
	for {
		rawStruct := <-channel
		notesChan <- rawStruct
	}
}

func sendNotesClient(wg *sync.WaitGroup, closedChan chan struct{}, conn net.Conn, delay int, notesChan chan interface{}, csvRecords []types.MidiCSVRecord, dontControlVolume bool) {

	var t *timer.Timer
	if !dontControlVolume {
		t = &timer.Timer{}
		timeout := t.New(10) // 10 seconds
		t.Start()

		go func() {
			// reset the volume on the timeout
			for range timeout {
				volume.SetVolume(common.HIGH_VOLUME)
				// make a new timer and overwrite the channel
				t = &timer.Timer{}
				timeout = t.New(10) // 10 seconds
			}
		}()
	}

	encoder := gob.NewEncoder(conn)
	reconnect := false
	for !reconnect {
		var msg interface{}
		select {
		case msg = <-notesChan:
		case <-closedChan:
			reconnect = true
			continue
		}

		if !dontControlVolume {
			volume.SetVolume(common.LOW_VOLUME)
			t.Reset()
		}

		go func() {
			if delay > 0 {
				time.Sleep(time.Duration(delay) * time.Millisecond)
			}
			// process messages differently based on type
			// this is just so we can deal with a single known struct with exposed fields
			switch v := msg.(type) {
			case channel.NoteOn:
				channel := channels.CheckChannel(v.Channel(), csvRecords)
				key := channels.CheckOffset(v.Channel(), v.Key(), csvRecords)
				if channel != 255 {
					err := encoder.Encode(types.TCPMessage{Body: types.NoteOn{
						Time:     time.Now(),
						Channel:  channel,
						Key:      key,
						Velocity: v.Velocity(),
					}})
					if err != nil {
						common.Cont(err)
						// put the note back on the channel
						notesChan <- msg
						reconnect = true
					}
				}
			case channel.NoteOff:
				channel := channels.CheckChannel(v.Channel(), csvRecords)
				key := channels.CheckOffset(v.Channel(), v.Key(), csvRecords)
				if channel != 255 {
					err := encoder.Encode(types.TCPMessage{Body: types.NoteOff{
						Time:    time.Now(),
						Channel: channel,
						Key:     key,
					}})
					if err != nil {
						common.Cont(err)
						notesChan <- msg
						reconnect = true
					}
				}
			case channel.ProgramChange:
				channel := channels.CheckChannel(v.Channel(), csvRecords)
				if channel != 255 {
					err := encoder.Encode(types.TCPMessage{Body: types.ProgramChange{
						Time:    time.Now(),
						Channel: channel,
						Program: v.Program(),
					}})
					if err != nil {
						common.Cont(err)
						notesChan <- msg
						reconnect = true
					}
				}
			case channel.Aftertouch:
				channel := channels.CheckChannel(v.Channel(), csvRecords)
				if channel != 255 {
					err := encoder.Encode(types.TCPMessage{Body: types.Aftertouch{
						Time:     time.Now(),
						Channel:  channel,
						Pressure: v.Pressure(),
					}})
					if err != nil {
						common.Cont(err)
						notesChan <- msg
						reconnect = true
					}
				}

			case channel.ControlChange:
				channel := channels.CheckChannel(v.Channel(), csvRecords)
				if channel != 255 {
					err := encoder.Encode(types.TCPMessage{Body: types.ControlChange{
						Time:       time.Now(),
						Channel:    channel,
						Controller: v.Controller(),
						Value:      v.Value(),
					}})
					if err != nil {
						common.Cont(err)
						notesChan <- msg
						reconnect = true
					}
				}
			case channel.NoteOffVelocity:
				channel := channels.CheckChannel(v.Channel(), csvRecords)
				key := channels.CheckOffset(v.Channel(), v.Key(), csvRecords)
				if channel != 255 {
					err := encoder.Encode(types.TCPMessage{Body: types.NoteOffVelocity{
						Time:     time.Now(),
						Channel:  channel,
						Key:      key,
						Velocity: v.Velocity(),
					}})
					if err != nil {
						common.Cont(err)
						notesChan <- msg
						reconnect = true
					}
				}
			case channel.Pitchbend:
				channel := channels.CheckChannel(v.Channel(), csvRecords)
				if channel != 255 {
					err := encoder.Encode(types.TCPMessage{Body: types.Pitchbend{
						Time:     time.Now(),
						Channel:  channel,
						Value:    v.Value(),
						AbsValue: v.AbsValue(),
					}})
					if err != nil {
						common.Cont(err)
						notesChan <- msg
						reconnect = true
					}
				}
			case channel.PolyAftertouch:
				channel := channels.CheckChannel(v.Channel(), csvRecords)
				key := channels.CheckOffset(v.Channel(), v.Key(), csvRecords)
				if channel != 255 {
					err := encoder.Encode(types.TCPMessage{Body: types.PolyAftertouch{
						Time:     time.Now(),
						Channel:  channel,
						Key:      key,
						Pressure: v.Pressure(),
					}})
					if err != nil {
						common.Cont(err)
						notesChan <- msg
						reconnect = true
					}
				}
			case types.Raw:
				err := encoder.Encode(types.TCPMessage{Body: types.Raw{
					Time: v.Time,
					Data: v.Data,
				}})
				if err != nil {
					common.Cont(err)
					notesChan <- msg
					reconnect = true
				}
			default:
				log.Println("Unknown message type:", v)
			}
		}()
	}
	wg.Done()
}

func midiClient(midiPort int, delay int, notesChan chan interface{}, in midi.In) {

	// listen for MIDI messages
	rd := reader.New(
		reader.NoLogger(),
		// write every message to the out port
		reader.Each(func(pos *reader.Position, msg midi.Message) {
			// send each message through the channel
			notesChan <- msg
		}),
	)
	common.Must(rd.ListenTo(in))
	// sleep forever
	select {}
}

// Listen for midi notes coming back so they can be printed
func midiClientFeedback(wg *sync.WaitGroup, closedChan chan struct{}, conn net.Conn, writers []*writer.Writer, out midi.Out, midiTuxChan chan types.MidiTuxMessage) {

	var t types.TCPMessage
	dec := gob.NewDecoder(conn)

	for {
		err := dec.Decode(&t)
		if err == io.EOF {
			log.Println("Connection closed by server.")
			closedChan <- struct{}{}
			wg.Done()
			return
		}
		if err != nil {
			common.Cont(err)
		} else {
			// print with midiTux
			switch m := t.Body.(type) {
			case types.NoteOn:
				ms := common.HandleMs(m.Time)
				common.Cont(writer.NoteOn(writers[m.Channel], m.Key, m.Velocity))
				midiTuxChan <- types.MidiTuxMessage{
					Color: color.FgHiGreen,
					T:     t.Body,
					Ms:    ms,
				}
			case types.NoteOff:
				ms := common.HandleMs(m.Time)
				common.Cont(writer.NoteOff(writers[m.Channel], m.Key))
				midiTuxChan <- types.MidiTuxMessage{
					Color: color.FgHiRed,
					T:     t.Body,
					Ms:    ms,
				}
			case types.ProgramChange:
				ms := common.HandleMs(m.Time)
				common.Cont(writer.ProgramChange(writers[m.Channel], m.Program))
				midiTuxChan <- types.MidiTuxMessage{
					Color: color.FgHiYellow,
					T:     t.Body,
					Ms:    ms,
				}
			case types.Aftertouch:
				ms := common.HandleMs(m.Time)
				common.Cont(writer.Aftertouch(writers[m.Channel], m.Pressure))
				midiTuxChan <- types.MidiTuxMessage{
					Color: color.FgHiBlue,
					T:     t.Body,
					Ms:    ms,
				}
			case types.ControlChange:
				ms := common.HandleMs(m.Time)
				common.Cont(writer.ControlChange(writers[m.Channel], m.Controller, m.Value))
				midiTuxChan <- types.MidiTuxMessage{
					Color: color.FgHiMagenta,
					T:     t.Body,
					Ms:    ms,
				}
			case types.NoteOffVelocity:
				ms := common.HandleMs(m.Time)
				common.Cont(writer.NoteOffVelocity(writers[m.Channel], m.Key, m.Velocity))
				midiTuxChan <- types.MidiTuxMessage{
					Color: color.FgYellow,
					T:     t.Body,
					Ms:    ms,
				}
			case types.Pitchbend:
				ms := common.HandleMs(m.Time)
				common.Cont(writer.Pitchbend(writers[m.Channel], m.Value))
				midiTuxChan <- types.MidiTuxMessage{
					Color: color.FgMagenta,
					T:     t.Body,
					Ms:    ms,
				}
			case types.PolyAftertouch:
				ms := common.HandleMs(m.Time)
				common.Cont(writer.PolyAftertouch(writers[m.Channel], m.Key, m.Pressure))
				midiTuxChan <- types.MidiTuxMessage{
					Color: color.FgCyan,
					T:     t.Body,
					Ms:    ms,
				}
			case types.Raw:
				ms := common.HandleMs(m.Time)
				if common.CheckAllNotesOff(m.Data) {
					// all notes off expansion
					common.ExpandAllNotesOff(m, ms, midiTuxChan, out)
				} else {
					// write the raw bytes to the MIDI device
					_, err := out.Write(m.Data)
					common.Cont(err)
				}
				midiTuxChan <- types.MidiTuxMessage{
					Color: color.FgBlue,
					T:     t.Body,
					Ms:    ms,
				}
			default:
				log.Println("Unknown message type:", m)
			}
		}
	}
}
