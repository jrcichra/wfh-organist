package main

import (
	"encoding/gob"
	"io"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/fatih/color"
	"gitlab.com/gomidi/midi/writer"
	driver "gitlab.com/gomidi/rtmididrv"
)

func server(midiPort int, serverPort int, protocol string) {
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
	l, err := net.Listen(protocol, ":"+strconv.Itoa(serverPort))
	must(err)
	defer l.Close()

	// keep accepting connections
	for {
		log.Println("Listening on", l.Addr())
		c, err := l.Accept()
		must(err)
		log.Println("Connection from:", c.RemoteAddr())
		log.Println("Ready to play music!")
		// handle the connection in a new goroutine
		go func() {
			// will read from network.
			dec := gob.NewDecoder(c)
			// gob.Register(ProgramChange{})

			// handle TCP messages forever
			for {
				var t TCPMessage
				err := dec.Decode(&t)
				if err == io.EOF {
					log.Println("Connection closed by client.")
					return
				}
				must(err)

				// determine the type of message
				switch m := t.Body.(type) {
				case NoteOn:
					ms := time.Since(m.Time).Milliseconds()
					cont(writer.NoteOn(writers[m.Channel], m.Key, m.Velocity))
					midiTuxPrint(color.FgHiGreen, c.RemoteAddr(), m, ms)
				case NoteOff:
					ms := time.Since(m.Time).Milliseconds()
					cont(writer.NoteOff(writers[m.Channel], m.Key))
					midiTuxPrint(color.FgHiRed, c.RemoteAddr(), m, ms)
				case ProgramChange:
					ms := time.Since(m.Time).Milliseconds()
					cont(writer.ProgramChange(writers[m.Channel], m.Program))
					midiTuxPrint(color.FgHiYellow, c.RemoteAddr(), m, ms)
				case Aftertouch:
					ms := time.Since(m.Time).Milliseconds()
					cont(writer.Aftertouch(writers[m.Channel], m.Pressure))
					midiTuxPrint(color.FgHiBlue, c.RemoteAddr(), m, ms)
				case ControlChange:
					ms := time.Since(m.Time).Milliseconds()
					cont(writer.ControlChange(writers[m.Channel], m.Controller, m.Value))
					midiTuxPrint(color.FgHiMagenta, c.RemoteAddr(), m, ms)
				case NoteOffVelocity:
					ms := time.Since(m.Time).Milliseconds()
					cont(writer.NoteOffVelocity(writers[m.Channel], m.Key, m.Velocity))
					midiTuxPrint(color.FgHiMagenta, c.RemoteAddr(), m, ms)
				case Pitchbend:
					ms := time.Since(m.Time).Milliseconds()
					cont(writer.Pitchbend(writers[m.Channel], m.Value))
					midiTuxPrint(color.FgHiMagenta, c.RemoteAddr(), m, ms)
				case PolyAftertouch:
					ms := time.Since(m.Time).Milliseconds()
					cont(writer.PolyAftertouch(writers[m.Channel], m.Key, m.Pressure))
					midiTuxPrint(color.FgHiMagenta, c.RemoteAddr(), m, ms)
				default:
					log.Println("Unknown message type:", m)
				}

			}
		}()
	}
}
