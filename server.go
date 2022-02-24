package main

import (
	"encoding/gob"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"

	"github.com/fatih/color"
	"gitlab.com/gomidi/midi/writer"
	driver "gitlab.com/gomidi/rtmididrv"
)

func startHTTP(notesChan chan interface{}) {
	// serve the website
	http.Handle("/", http.FileServer(http.Dir("./gui/dist")))
	//serve favicon
	http.Handle("/favicon.ico", http.FileServer(http.Dir("./gui/build/favicon.ico")))
	// serve /api
	http.Handle("/api/midi/raw", handleAPI(notesChan))
	// http listener
	log.Println("HTTP Listening on 8080")
	http.ListenAndServe(":8080", nil)
}

func server(midiPort int, serverPort int, protocol string, midiTuxChan chan MidiTuxMessage) {

	// wait for someone to connect to the server
	l, err := net.Listen(protocol, ":"+strconv.Itoa(serverPort))
	must(err)
	defer l.Close()

	//send notes listening to a go channel
	notesChan := make(chan interface{})
	go sendNotes(midiPort, notesChan, midiTuxChan)

	// send notes back to the client too
	feedbackChan := make(chan interface{})
	go feedbackNotes(feedbackChan)

	// also can accept notes from the HTTP API
	go startHTTP(notesChan)

	// keep accepting connections
	for {
		log.Println("Notes listening on", l.Addr())
		c, err := l.Accept()
		must(err)
		log.Println("Notes connection from:", c.RemoteAddr())
		log.Println("Ready to play music!")

		go func() {
			dec := gob.NewDecoder(c)
			for {
				var t TCPMessage
				err := dec.Decode(&t)
				if err == io.EOF {
					log.Println("Connection closed by client.")
					feedbackChan <- nil
					c.Close()
					return
				}
				must(err)
				// send through the channel
				notesChan <- t.Body
				// send back through to the client
				feedbackChan <- t.Body
			}
		}()
	}
}

// send notes back to the client from the server
func feedbackNotes(feedbackChan chan interface{}) {
	// listen for clients on 3132
	l, err := net.Listen("tcp", ":3132")
	must(err)
	defer l.Close()
	for {
		log.Println("Feedback Listening on", l.Addr())
		// accept user
		c, err := l.Accept()
		must(err)
		log.Println("Feedback connection from:", c.RemoteAddr())
		go func() {
			encoder := gob.NewEncoder(c)
			for {
				note := <-feedbackChan
				if note == nil {
					log.Println("Feedback connection closed by client.")
					c.Close()
					return
				}
				err := encoder.Encode(TCPMessage{Body: note})
				if err != nil {
					log.Println(err)
					c.Close()
					return
				}
			}
		}()
	}
}

func sendNotes(midiPort int, notesChan chan interface{}, midiTuxChan chan MidiTuxMessage) {

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

	for {
		input := <-notesChan
		// determine the type of message
		switch m := input.(type) {
		case NoteOn:
			ms := handleMs(m.Time)
			cont(writer.NoteOn(writers[m.Channel], m.Key, m.Velocity))
			midiTuxChan <- MidiTuxMessage{
				Color: color.FgHiGreen,
				T:     m,
				Ms:    ms,
			}
		case NoteOff:
			ms := handleMs(m.Time)
			cont(writer.NoteOff(writers[m.Channel], m.Key))
			midiTuxChan <- MidiTuxMessage{
				Color: color.FgHiRed,
				T:     m,
				Ms:    ms,
			}
		case ProgramChange:
			ms := handleMs(m.Time)
			cont(writer.ProgramChange(writers[m.Channel], m.Program))
			midiTuxChan <- MidiTuxMessage{
				Color: color.FgHiYellow,
				T:     m,
				Ms:    ms,
			}
		case Aftertouch:
			ms := handleMs(m.Time)
			cont(writer.Aftertouch(writers[m.Channel], m.Pressure))
			midiTuxChan <- MidiTuxMessage{
				Color: color.FgHiBlue,
				T:     m,
				Ms:    ms,
			}
		case ControlChange:
			ms := handleMs(m.Time)
			cont(writer.ControlChange(writers[m.Channel], m.Controller, m.Value))
			midiTuxChan <- MidiTuxMessage{
				Color: color.FgHiMagenta,
				T:     m,
				Ms:    ms,
			}
		case NoteOffVelocity:
			ms := handleMs(m.Time)
			cont(writer.NoteOffVelocity(writers[m.Channel], m.Key, m.Velocity))
			midiTuxChan <- MidiTuxMessage{
				Color: color.FgYellow,
				T:     m,
				Ms:    ms,
			}
		case Pitchbend:
			ms := handleMs(m.Time)
			cont(writer.Pitchbend(writers[m.Channel], m.Value))
			midiTuxChan <- MidiTuxMessage{
				Color: color.FgMagenta,
				T:     m,
				Ms:    ms,
			}
		case PolyAftertouch:
			ms := handleMs(m.Time)
			cont(writer.PolyAftertouch(writers[m.Channel], m.Key, m.Pressure))
			midiTuxChan <- MidiTuxMessage{
				Color: color.FgCyan,
				T:     m,
				Ms:    ms,
			}
		case Raw:
			ms := handleMs(m.Time)
			midiTuxPrint(color.FgBlue, m, ms)
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
				T:     m,
				Ms:    ms,
			}
		default:
			log.Println("Unknown message type:", m)
		}
	}
}
