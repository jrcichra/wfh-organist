package main

import (
	"bufio"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
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

func client(midiPort int, serverIP string, serverPort int, protocol string, stdinMode bool) {

	switch stdinMode {
	case true:
		stdinClient(serverIP, serverPort, protocol)
	default:
		midiClient(midiPort, serverIP, serverPort, protocol)
	}
}

func stdinClient(serverIP string, serverPort int, protocol string) {

	channel := make(chan Raw)
	//get stdin in a goroutine
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			// read all the tokens in this line and split them by whitespace
			tokens := strings.Fields(scanner.Text())
			for _, token := range tokens {
				//convert token string to hex code
				log.Println("Received:", token)
				hexToken, err := hex.DecodeString(token)
				must(err)
				//send hex code to channel
				log.Println("Sending to channel:", token)
				channel <- Raw{
					Time: time.Now(),
					Data: hexToken,
				}
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
	log.Println("Listening to channel:")
	for {
		rawStruct := <-channel
		log.Println("Got", rawStruct.Data, "from channel")
		err := encoder.Encode(TCPMessage{Body: rawStruct}) // sends a Raw struct to the server
		must(err)
	}
}

func midiClient(midiPort int, serverIP string, serverPort int, protocol string) {

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
				// process messages differently based on type
				// this is just so we can deal with a single known struct with exposed fields
				switch v := msg.(type) {
				case channel.NoteOn:
					err := encoder.Encode(TCPMessage{Body: NoteOn{
						Time:     time.Now(),
						Channel:  v.Channel(),
						Key:      v.Key(),
						Velocity: v.Velocity(),
					}})
					must(err)
				case channel.NoteOff:
					err := encoder.Encode(TCPMessage{Body: NoteOff{
						Time:    time.Now(),
						Channel: v.Channel(),
						Key:     v.Key(),
					}})
					must(err)
				case channel.ProgramChange:
					err := encoder.Encode(TCPMessage{Body: ProgramChange{
						Time:    time.Now(),
						Channel: v.Channel(),
						Program: v.Program(),
					}})
					must(err)
				case channel.Aftertouch:
					err := encoder.Encode(TCPMessage{Body: Aftertouch{
						Time:     time.Now(),
						Channel:  v.Channel(),
						Pressure: v.Pressure(),
					}})
					must(err)
				case channel.ControlChange:
					err := encoder.Encode(TCPMessage{Body: ControlChange{
						Time:       time.Now(),
						Channel:    v.Channel(),
						Controller: v.Controller(),
						Value:      v.Value(),
					}})
					must(err)
				case channel.NoteOffVelocity:
					err := encoder.Encode(TCPMessage{Body: NoteOffVelocity{
						Time:     time.Now(),
						Channel:  v.Channel(),
						Key:      v.Key(),
						Velocity: v.Velocity(),
					}})
					must(err)
				case channel.Pitchbend:
					err := encoder.Encode(TCPMessage{Body: Pitchbend{
						Time:     time.Now(),
						Channel:  v.Channel(),
						Value:    v.Value(),
						AbsValue: v.AbsValue(),
					}})
					must(err)
				case channel.PolyAftertouch:
					err := encoder.Encode(TCPMessage{Body: PolyAftertouch{
						Time:     time.Now(),
						Channel:  v.Channel(),
						Key:      v.Key(),
						Pressure: v.Pressure(),
					}})
					must(err)
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
