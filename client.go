package main

import (
	"bufio"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"gitlab.com/gomidi/midi"
	"gitlab.com/gomidi/midi/midimessage/channel"
	"gitlab.com/gomidi/midi/reader"
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

func client(midiPort int, serverIP string, serverPort int, protocol string, stdinMode bool, delay int, csvRecords []MidiCSVRecord) {

	switch stdinMode {
	case true:
		stdinClient(serverIP, serverPort, protocol)
	default:
		midiClient(midiPort, serverIP, serverPort, protocol, delay, csvRecords)
	}
}

func stdinClient(serverIP string, serverPort int, protocol string) {

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

	//send stdin to server
	conn := dial(serverIP, serverPort, protocol)
	defer conn.Close()
	// prepare to encode raw
	encoder := gob.NewEncoder(conn)
	// read from the channel and send to server
	for {
		rawStruct := <-channel
		err := encoder.Encode(TCPMessage{Body: rawStruct}) // sends a Raw struct to the server
		must(err)
	}
}

func midiClient(midiPort int, serverIP string, serverPort int, protocol string, delay int, csvRecords []MidiCSVRecord) {

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

	conn := dial(serverIP, serverPort, protocol)
	encoder := gob.NewEncoder(conn)

	// listen for MIDI messages
	rd := reader.New(
		reader.NoLogger(),
		// write every message to the out port
		reader.Each(func(pos *reader.Position, msg midi.Message) {
			// send each message in a separate goroutine
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
				default:
					log.Println("Unknown message type:", v)
				}
			}()
		}),
	)
	must(rd.ListenTo(in))
	// sleep forever
	select {}
}
