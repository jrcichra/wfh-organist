package server

import (
	"encoding/gob"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"

	"github.com/fatih/color"
	"github.com/jrcichra/wfh-organist/internal/common"
	"github.com/jrcichra/wfh-organist/internal/parser/config"
	"github.com/jrcichra/wfh-organist/internal/recorder"
	"github.com/jrcichra/wfh-organist/internal/state"
	"github.com/jrcichra/wfh-organist/internal/types"
	"gitlab.com/gomidi/midi"
	"gitlab.com/gomidi/midi/writer"
	driver "gitlab.com/gomidi/rtmididrv"
)

type Server struct {
	Profile     string
	MidiPort    int
	Port        int
	DontRecord  bool
	state       *state.State
	notesChan   chan interface{}
	stops       *config.Config
	MidiTuxChan chan types.MidiTuxMessage
	out         midi.Out
	in          midi.In
}

func (s *Server) startHTTP() {
	// serve the website
	http.Handle("/", http.FileServer(http.Dir("./gui/dist")))
	//serve favicon
	http.Handle("/favicon.ico", http.FileServer(http.Dir("./gui/build/favicon.ico")))
	// serve /api
	http.Handle("/api/midi/", s.handleAPI())
	// http listener
	log.Println("HTTP Listening on 8080")
	http.ListenAndServe(":8080", nil)
}

func (s *Server) Run() {

	// wait for someone to connect to the server
	l, err := net.Listen("tcp", ":"+strconv.Itoa(s.Port))
	common.Must(err)
	defer l.Close()

	drv, err := driver.New()
	common.Must(err)
	// make sure to close all open ports at the end
	defer drv.Close()

	s.out = common.GetMidiOutput(drv, s.MidiPort)

	//send notes listening to a go channel
	s.notesChan = make(chan interface{})
	go s.sendNotes()

	// record to a file
	if !s.DontRecord {
		s.in = common.GetMidiInput(drv, s.MidiPort)
		stopRecording := make(chan bool)
		common.SetupCloseHandler(s.out, stopRecording)
		go recorder.Record(s.in, stopRecording)
	}

	s.state = &state.State{}
	s.state.Open(s.Profile, s.notesChan)

	// also can accept notes from the HTTP API
	go s.startHTTP()

	// keep accepting connections
	for {
		log.Println("Notes listening on", l.Addr())
		c, err := l.Accept()
		common.Must(err)
		log.Println("Notes connection from:", c.RemoteAddr())
		log.Println("Ready to play music!")

		go func() {
			dec := gob.NewDecoder(c)
			enc := gob.NewEncoder(c)
			for {
				var t types.TCPMessage
				err := dec.Decode(&t)
				if err == io.EOF {
					log.Println("Connection closed by client.")
					c.Close()
					return
				}
				common.Must(err)
				// send through the channel
				s.notesChan <- t.Body
				// and send it through feedback channel
				err = enc.Encode(types.TCPMessage{Body: t.Body})
				if err != nil {
					log.Println(err)
				}
			}
		}()
	}
}

func (s *Server) sendNotes() {

	// make a writer for each channel
	writers := make([]*writer.Writer, 16)
	var i uint8
	for ; i < 16; i++ {
		writers[i] = writer.New(s.out)
		writers[i].SetChannel(i)
	}

	for {
		input := <-s.notesChan
		// determine the type of message
		switch m := input.(type) {
		case types.NoteOn:
			ms := common.HandleMs(m.Time)
			common.Cont(writer.NoteOn(writers[m.Channel], m.Key, m.Velocity))
			s.MidiTuxChan <- types.MidiTuxMessage{
				Color: color.FgHiGreen,
				T:     m,
				Ms:    ms,
			}
		case types.NoteOff:
			ms := common.HandleMs(m.Time)
			common.Cont(writer.NoteOff(writers[m.Channel], m.Key))
			s.MidiTuxChan <- types.MidiTuxMessage{
				Color: color.FgHiRed,
				T:     m,
				Ms:    ms,
			}
		case types.ProgramChange:
			ms := common.HandleMs(m.Time)
			common.Cont(writer.ProgramChange(writers[m.Channel], m.Program))
			s.MidiTuxChan <- types.MidiTuxMessage{
				Color: color.FgHiYellow,
				T:     m,
				Ms:    ms,
			}
		case types.Aftertouch:
			ms := common.HandleMs(m.Time)
			common.Cont(writer.Aftertouch(writers[m.Channel], m.Pressure))
			s.MidiTuxChan <- types.MidiTuxMessage{
				Color: color.FgHiBlue,
				T:     m,
				Ms:    ms,
			}
		case types.ControlChange:
			ms := common.HandleMs(m.Time)
			common.Cont(writer.ControlChange(writers[m.Channel], m.Controller, m.Value))
			s.MidiTuxChan <- types.MidiTuxMessage{
				Color: color.FgHiMagenta,
				T:     m,
				Ms:    ms,
			}
		case types.NoteOffVelocity:
			ms := common.HandleMs(m.Time)
			common.Cont(writer.NoteOffVelocity(writers[m.Channel], m.Key, m.Velocity))
			s.MidiTuxChan <- types.MidiTuxMessage{
				Color: color.FgYellow,
				T:     m,
				Ms:    ms,
			}
		case types.Pitchbend:
			ms := common.HandleMs(m.Time)
			common.Cont(writer.Pitchbend(writers[m.Channel], m.Value))
			s.MidiTuxChan <- types.MidiTuxMessage{
				Color: color.FgMagenta,
				T:     m,
				Ms:    ms,
			}
		case types.PolyAftertouch:
			ms := common.HandleMs(m.Time)
			common.Cont(writer.PolyAftertouch(writers[m.Channel], m.Key, m.Pressure))
			s.MidiTuxChan <- types.MidiTuxMessage{
				Color: color.FgCyan,
				T:     m,
				Ms:    ms,
			}
		case types.Raw:
			ms := common.HandleMs(m.Time)
			s.MidiTuxChan <- types.MidiTuxMessage{
				Color: color.FgHiBlue,
				T:     m,
				Ms:    ms,
			}
			if common.CheckAllNotesOff(m.Data) {
				// all notes off expansion
				common.ExpandAllNotesOff(m, ms, s.MidiTuxChan, s.out)
			} else {
				// write the raw bytes to the MIDI device
				_, err := s.out.Write(m.Data)
				common.Cont(err)
			}
		default:
			log.Println("Unknown message type:", m)
		}
	}
}
