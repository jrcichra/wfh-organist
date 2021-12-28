package main

import (
	"encoding/gob"
	"log"
	"net"
	"time"

	"github.com/fatih/color"
	"gitlab.com/gomidi/midi/writer"
	driver "gitlab.com/gomidi/rtmididrv"
)

func server(midiPort int) {
	drv, err := driver.New()
	must(err)

	// make sure to close all open ports at the end
	defer drv.Close()

	outs, err := drv.Outs()
	must(err)

	if len(outs)-1 < midiPort {
		log.Printf("Too few MIDI OUT Ports found. Wanted Index: %d. Max Index: %d\n", midiPort, len(outs)-1)
		return
	}
	out := outs[midiPort]

	must(out.Open())

	// make a writer for each channel
	writers := make([]*writer.Writer, 16)
	var i uint8
	for ; i < 16; i++ {
		writers[i] = writer.New(out)
		writers[i].SetChannel(i)
	}

	// wait for someone to connect to the server
	l, err := net.Listen("tcp", ":3131")
	must(err)
	defer l.Close()

	// keep accepting connections
	for {
		log.Println("Listening on", l.Addr())
		c, err := l.Accept()
		must(err)
		log.Println("Connection from:", c.RemoteAddr())

		// handle the connection in a new goroutine
		go func() {
			// will read from network.
			dec := gob.NewDecoder(c)
			// handle TCP messages forever
			for {
				t := TCPMessage{}
				dec.Decode(&t)
				// print time delay in ms
				ms := time.Since(t.Time).Milliseconds()

				// NoteOff = Velocity 0
				if t.Velocity == 0 {
					cont(writer.NoteOff(writers[t.Channel], t.Key))
					midiTuxPrint(color.FgHiRed, c.RemoteAddr(), t, ms)
				} else {
					cont(writer.NoteOn(writers[t.Channel], t.Key, t.Velocity))
					midiTuxPrint(color.FgHiGreen, c.RemoteAddr(), t, ms)
				}
			}
		}()
	}
}
