package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"time"

	"gitlab.com/gomidi/midi"
	"gitlab.com/gomidi/midi/midimessage/channel"
	"gitlab.com/gomidi/midi/reader"
	driver "gitlab.com/gomidi/rtmididrv"
)

func client(midiPort int, serverIP string, serverPort int) {

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

	serverStr := fmt.Sprintf("%s:%d", serverIP, serverPort)
	log.Println("Connecting to " + serverStr + "...")
	conn, err := net.Dial("tcp", serverStr)
	must(err)
	log.Println("Connected to", serverStr)

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
					encoder.Encode(TCPMessage{
						Time:     time.Now(),
						Channel:  v.Channel(),
						Key:      v.Key(),
						Velocity: v.Velocity(),
					})
				case channel.NoteOff:
					encoder.Encode(TCPMessage{
						Time:     time.Now(),
						Channel:  v.Channel(),
						Key:      v.Key(),
						Velocity: 0,
					})
				}
			}()
		}),
	)
	must(rd.ListenTo(in))
	// sleep forever
	select {}
}
