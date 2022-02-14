package main

import (
	"bufio"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

	"github.com/fatih/color"
	"gitlab.com/gomidi/midi"
	"gitlab.com/gomidi/midi/midimessage/channel"
	"gitlab.com/gomidi/midi/reader"
	"gitlab.com/gomidi/midi/writer"
	driver "gitlab.com/gomidi/rtmididrv"
)

func dial(serverIP string, serverPort int, protocol string) net.Conn {
	serverStr := fmt.Sprintf("%s:%d", serverIP, serverPort)
	log.Println("Connecting to " + serverStr + "...")
	conn, err := net.Dial(protocol, serverStr)
	must(err)
	log.Println("Connected to", serverStr)
	return conn
}

func client(midiPort int, serverIP string, serverPort int, protocol string, stdinMode bool, delay int, file string, midiTuxChan chan MidiTuxMessage) {

	// read the csv
	csvRecords := readCSV()

	notesChan := make(chan interface{})

	drv, err := driver.New()
	must(err)

	// make sure to close all open ports at the end
	defer drv.Close()

	ins, err := drv.Ins()
	must(err)

	if len(ins)-1 < midiPort {
		log.Printf("Too few MIDI IN Ports found. Wanted Index: %d. Max Index: %d\n", midiPort, len(ins)-1)
		return
	}
	in := ins[midiPort]

	must(in.Open())

	outs, err := drv.Outs()
	must(err)

	out := outs[midiPort]

	must(out.Open())

	// make a writer for each channel
	writers := make([]*writer.Writer, 16)
	var i uint8
	for ; i < 16; i++ {
		writers[i] = writer.New(out)
		writers[i].SetChannel(i)
	}

	// in either mode read the serial for now
	go readSerial(notesChan)
	// ability to send notes
	go sendNotesClient(serverIP, serverPort, protocol, delay, notesChan, csvRecords)
	// ability to get your own notes back
	go midiClientFeedback(serverIP, 3132, protocol, writers, out, midiTuxChan)
	switch stdinMode {
	case true:
		stdinClient(serverIP, serverPort, protocol, notesChan)
	default:
		switch file == "" {
		case true:
			midiClient(midiPort, delay, notesChan, in)
		default:
			playMidiFile(notesChan, file)
		}
	}
}

func stdinClient(serverIP string, serverPort int, protocol string, notesChan chan interface{}) {

	channel := make(chan Raw)

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
				panic("Token must be size 2")
			}
			hexToken, err := hex.DecodeString(text)
			must(err)
			// append to bytes
			bytes = append(bytes, hexToken...)
			if count >= 2 {
				//send hex code to channel
				channel <- Raw{
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

func sendNotesClient(serverIP string, serverPort int, protocol string, delay int, notesChan chan interface{}, csvRecords []MidiCSVRecord) {

	conn := dial(serverIP, serverPort, protocol)
	encoder := gob.NewEncoder(conn)

	for {
		msg := <-notesChan
		go func() {
			if delay > 0 {
				time.Sleep(time.Duration(delay) * time.Millisecond)
			}
			// process messages differently based on type
			// this is just so we can deal with a single known struct with exposed fields
			switch v := msg.(type) {
			case channel.NoteOn:
				channel := csvCheckChannel(v.Channel(), csvRecords)
				key := csvCheckOffset(v.Channel(), v.Key(), csvRecords)
				if channel != 255 {
					err := encoder.Encode(TCPMessage{Body: NoteOn{
						Time:     time.Now(),
						Channel:  channel,
						Key:      key,
						Velocity: v.Velocity(),
					}})
					must(err)
				}
			case channel.NoteOff:
				channel := csvCheckChannel(v.Channel(), csvRecords)
				key := csvCheckOffset(v.Channel(), v.Key(), csvRecords)
				if channel != 255 {
					err := encoder.Encode(TCPMessage{Body: NoteOff{
						Time:    time.Now(),
						Channel: channel,
						Key:     key,
					}})
					must(err)
				}
			case channel.ProgramChange:
				channel := csvCheckChannel(v.Channel(), csvRecords)
				if channel != 255 {
					err := encoder.Encode(TCPMessage{Body: ProgramChange{
						Time:    time.Now(),
						Channel: channel,
						Program: v.Program(),
					}})
					must(err)
				}
			case channel.Aftertouch:
				channel := csvCheckChannel(v.Channel(), csvRecords)
				if channel != 255 {
					err := encoder.Encode(TCPMessage{Body: Aftertouch{
						Time:     time.Now(),
						Channel:  channel,
						Pressure: v.Pressure(),
					}})
					must(err)
				}

			case channel.ControlChange:
				channel := csvCheckChannel(v.Channel(), csvRecords)
				if channel != 255 {
					err := encoder.Encode(TCPMessage{Body: ControlChange{
						Time:       time.Now(),
						Channel:    channel,
						Controller: v.Controller(),
						Value:      v.Value(),
					}})
					must(err)
				}
			case channel.NoteOffVelocity:
				channel := csvCheckChannel(v.Channel(), csvRecords)
				key := csvCheckOffset(v.Channel(), v.Key(), csvRecords)
				if channel != 255 {
					err := encoder.Encode(TCPMessage{Body: NoteOffVelocity{
						Time:     time.Now(),
						Channel:  channel,
						Key:      key,
						Velocity: v.Velocity(),
					}})
					must(err)
				}
			case channel.Pitchbend:
				channel := csvCheckChannel(v.Channel(), csvRecords)
				if channel != 255 {
					err := encoder.Encode(TCPMessage{Body: Pitchbend{
						Time:     time.Now(),
						Channel:  channel,
						Value:    v.Value(),
						AbsValue: v.AbsValue(),
					}})
					must(err)
				}
			case channel.PolyAftertouch:
				channel := csvCheckChannel(v.Channel(), csvRecords)
				key := csvCheckOffset(v.Channel(), v.Key(), csvRecords)
				if channel != 255 {
					err := encoder.Encode(TCPMessage{Body: PolyAftertouch{
						Time:     time.Now(),
						Channel:  channel,
						Key:      key,
						Pressure: v.Pressure(),
					}})
					must(err)
				}
			case Raw:
				err := encoder.Encode(TCPMessage{Body: Raw{
					Time: v.Time,
					Data: v.Data,
				}})
				must(err)
			default:
				log.Println("Unknown message type:", v)
			}
		}()
	}
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
	must(rd.ListenTo(in))
	// sleep forever
	select {}
}

// Listen for midi notes coming back so they can be printed
func midiClientFeedback(serverIP string, serverPort int, protocol string, writers []*writer.Writer, out midi.Out, midiTuxChan chan MidiTuxMessage) {
	conn := dial(serverIP, serverPort, protocol)
	dec := gob.NewDecoder(conn)

	for {
		var t TCPMessage
		err := dec.Decode(&t)
		if err == io.EOF {
			log.Println("Feedback connection closed by server.")
			conn.Close()
			return
		}
		if err != nil {
			cont(err)
		} else {
			// print with midiTux
			switch m := t.Body.(type) {
			case NoteOn:
				ms := handleMs(m.Time)
				cont(writer.NoteOn(writers[m.Channel], m.Key, m.Velocity))
				midiTuxChan <- MidiTuxMessage{
					Color: color.FgHiGreen,
					T:     t.Body,
					Ms:    ms,
				}
			case NoteOff:
				ms := handleMs(m.Time)
				cont(writer.NoteOff(writers[m.Channel], m.Key))
				midiTuxChan <- MidiTuxMessage{
					Color: color.FgHiRed,
					T:     t.Body,
					Ms:    ms,
				}
			case ProgramChange:
				ms := handleMs(m.Time)
				cont(writer.ProgramChange(writers[m.Channel], m.Program))
				midiTuxChan <- MidiTuxMessage{
					Color: color.FgHiYellow,
					T:     t.Body,
					Ms:    ms,
				}
			case Aftertouch:
				ms := handleMs(m.Time)
				cont(writer.Aftertouch(writers[m.Channel], m.Pressure))
				midiTuxChan <- MidiTuxMessage{
					Color: color.FgHiBlue,
					T:     t.Body,
					Ms:    ms,
				}
			case ControlChange:
				ms := handleMs(m.Time)
				cont(writer.ControlChange(writers[m.Channel], m.Controller, m.Value))
				midiTuxChan <- MidiTuxMessage{
					Color: color.FgHiMagenta,
					T:     t.Body,
					Ms:    ms,
				}
			case NoteOffVelocity:
				ms := handleMs(m.Time)
				cont(writer.NoteOffVelocity(writers[m.Channel], m.Key, m.Velocity))
				midiTuxChan <- MidiTuxMessage{
					Color: color.FgYellow,
					T:     t.Body,
					Ms:    ms,
				}
			case Pitchbend:
				ms := handleMs(m.Time)
				cont(writer.Pitchbend(writers[m.Channel], m.Value))
				midiTuxChan <- MidiTuxMessage{
					Color: color.FgMagenta,
					T:     t.Body,
					Ms:    ms,
				}
			case PolyAftertouch:
				ms := handleMs(m.Time)
				cont(writer.PolyAftertouch(writers[m.Channel], m.Key, m.Pressure))
				midiTuxChan <- MidiTuxMessage{
					Color: color.FgCyan,
					T:     t.Body,
					Ms:    ms,
				}
			case Raw:
				ms := handleMs(m.Time)
				if checkAllNotesOff(m.Data) {
					// all notes off expansion
					expandAllNotesOff(m, ms, midiTuxChan, out)
				} else {
					// write the raw bytes to the MIDI device
					_, err := out.Write(m.Data)
					cont(err)
				}
				midiTuxChan <- MidiTuxMessage{
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
