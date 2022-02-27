package server

import (
	"encoding/gob"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"

	"github.com/fatih/color"
	"github.com/google/uuid"
	"github.com/jrcichra/wfh-organist/internal/common"
	"github.com/jrcichra/wfh-organist/internal/recorder"
	"github.com/jrcichra/wfh-organist/internal/types"
	"gitlab.com/gomidi/midi"
	"gitlab.com/gomidi/midi/writer"
	driver "gitlab.com/gomidi/rtmididrv"
)

func startHTTP(notesChan chan interface{}) {
	// serve the website
	http.Handle("/", http.FileServer(http.Dir("./gui/dist")))
	//serve favicon
	http.Handle("/favicon.ico", http.FileServer(http.Dir("./gui/build/favicon.ico")))
	// serve /api
	http.Handle("/api/midi/", handleAPI(notesChan))
	// http listener
	log.Println("HTTP Listening on 8080")
	http.ListenAndServe(":8080", nil)
}

func Server(midiPort int, serverPort int, protocol string, midiTuxChan chan types.MidiTuxMessage) {

	// wait for someone to connect to the server
	l, err := net.Listen(protocol, ":"+strconv.Itoa(serverPort))
	common.Must(err)
	defer l.Close()

	in, out := getMidiIO(midiPort)

	// send notes back to the client
	feedbackChan := make(chan interface{})
	go feedbackNotes(feedbackChan)

	//send notes listening to a go channel
	notesChan := make(chan interface{})
	go sendNotes(out, notesChan, midiTuxChan, feedbackChan)

	// always record to a file
	stopRecording := make(chan struct{})
	common.SetupCloseHandler(out, stopRecording)
	go recorder.Record(in, stopRecording)

	// also can accept notes from the HTTP API
	go startHTTP(notesChan)

	// keep accepting connections
	for {
		log.Println("Notes listening on", l.Addr())
		c, err := l.Accept()
		common.Must(err)
		log.Println("Notes connection from:", c.RemoteAddr())
		log.Println("Ready to play music!")

		go func() {
			dec := gob.NewDecoder(c)
			for {
				var t types.TCPMessage
				err := dec.Decode(&t)
				if err == io.EOF {
					log.Println("Connection closed by client.")
					feedbackChan <- nil
					c.Close()
					return
				}
				common.Must(err)
				// send through the channel
				notesChan <- t.Body
			}
		}()
	}
}

// send notes back to the client from the server
func feedbackNotes(feedbackChan chan interface{}) {
	// listen for clients on 3132
	l, err := net.Listen("tcp", ":3132")
	common.Must(err)
	defer l.Close()
	clients := make(map[string]net.Conn)
	encoders := make(map[string]*gob.Encoder)

	go func() {
		// get notes from the feedback chan and send them to every client we know about
		for note := range feedbackChan {
			for name, e := range encoders {
				err := e.Encode(types.TCPMessage{Body: note})
				if err != nil {
					log.Println(err)
					// If there was an error, close the connection and remove the client from the maps
					err := clients[name].Close()
					common.Cont(err)
					delete(clients, name)
					delete(encoders, name)
					return
				}
			}
		}
	}()

	for {
		log.Println("Feedback Listening on", l.Addr())
		// accept user
		c, err := l.Accept()
		common.Must(err)
		log.Println("Feedback connection from:", c.RemoteAddr())
		go func() {
			encoder := gob.NewEncoder(c)
			// add the client to the maps
			id := uuid.New().String()
			clients[id] = c
			encoders[id] = encoder
		}()
	}
}

func getMidiIO(midiPort int) (in midi.In, out midi.Out) {

	drv, err := driver.New()
	common.Must(err)
	// make sure to close all open ports at the end

	outs, err := drv.Outs()
	common.Must(err)

	ins, err := drv.Ins()
	common.Must(err)

	if len(outs)-1 < midiPort {
		log.Fatalf("Too few MIDI OUT Ports found. Wanted Index: %d. Max Index: %d\n", midiPort, len(outs)-1)
	}

	if len(ins)-1 < midiPort {
		log.Fatalf("Too few MIDI IN Ports found. Wanted Index: %d. Max Index: %d\n", midiPort, len(ins)-1)
	}

	out = outs[midiPort]
	common.Must(out.Open())

	in = ins[midiPort]
	common.Must(in.Open())
	return in, out
}

func sendNotes(out midi.Out, notesChan chan interface{}, midiTuxChan chan types.MidiTuxMessage, feedbackChan chan interface{}) {

	// make a writer for each channel
	writers := make([]*writer.Writer, 16)
	var i uint8
	for ; i < 16; i++ {
		writers[i] = writer.New(out)
		writers[i].SetChannel(i)
	}

	for {
		input := <-notesChan
		// send it out the feedback if someones there
		select {
		case feedbackChan <- input:
		default:
		}
		// determine the type of message
		switch m := input.(type) {
		case types.NoteOn:
			ms := common.HandleMs(m.Time)
			common.Cont(writer.NoteOn(writers[m.Channel], m.Key, m.Velocity))
			midiTuxChan <- types.MidiTuxMessage{
				Color: color.FgHiGreen,
				T:     m,
				Ms:    ms,
			}
		case types.NoteOff:
			ms := common.HandleMs(m.Time)
			common.Cont(writer.NoteOff(writers[m.Channel], m.Key))
			midiTuxChan <- types.MidiTuxMessage{
				Color: color.FgHiRed,
				T:     m,
				Ms:    ms,
			}
		case types.ProgramChange:
			ms := common.HandleMs(m.Time)
			common.Cont(writer.ProgramChange(writers[m.Channel], m.Program))
			midiTuxChan <- types.MidiTuxMessage{
				Color: color.FgHiYellow,
				T:     m,
				Ms:    ms,
			}
		case types.Aftertouch:
			ms := common.HandleMs(m.Time)
			common.Cont(writer.Aftertouch(writers[m.Channel], m.Pressure))
			midiTuxChan <- types.MidiTuxMessage{
				Color: color.FgHiBlue,
				T:     m,
				Ms:    ms,
			}
		case types.ControlChange:
			ms := common.HandleMs(m.Time)
			common.Cont(writer.ControlChange(writers[m.Channel], m.Controller, m.Value))
			midiTuxChan <- types.MidiTuxMessage{
				Color: color.FgHiMagenta,
				T:     m,
				Ms:    ms,
			}
		case types.NoteOffVelocity:
			ms := common.HandleMs(m.Time)
			common.Cont(writer.NoteOffVelocity(writers[m.Channel], m.Key, m.Velocity))
			midiTuxChan <- types.MidiTuxMessage{
				Color: color.FgYellow,
				T:     m,
				Ms:    ms,
			}
		case types.Pitchbend:
			ms := common.HandleMs(m.Time)
			common.Cont(writer.Pitchbend(writers[m.Channel], m.Value))
			midiTuxChan <- types.MidiTuxMessage{
				Color: color.FgMagenta,
				T:     m,
				Ms:    ms,
			}
		case types.PolyAftertouch:
			ms := common.HandleMs(m.Time)
			common.Cont(writer.PolyAftertouch(writers[m.Channel], m.Key, m.Pressure))
			midiTuxChan <- types.MidiTuxMessage{
				Color: color.FgCyan,
				T:     m,
				Ms:    ms,
			}
		case types.Raw:
			ms := common.HandleMs(m.Time)
			midiTuxChan <- types.MidiTuxMessage{
				Color: color.FgBlue,
				T:     m,
				Ms:    ms,
			}
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
				T:     m,
				Ms:    ms,
			}
		default:
			log.Println("Unknown message type:", m)
		}
	}
}
