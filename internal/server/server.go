package server

import (
	"context"
	"encoding/gob"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"sync"

	"github.com/fatih/color"
	"github.com/gorilla/websocket"
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
	Profile              string
	MidiPort             int
	Port                 int
	DontRecord           bool
	state                *state.State
	notesChan            chan interface{}
	feedbackChannels     map[string]chan interface{}
	feedbackChannelMutex sync.Mutex
	websockets           map[string]*websocket.Conn
	websocketsMutex      sync.Mutex
	stops                *config.Config
	MidiTuxChan          chan types.MidiTuxMessage
	out                  midi.Out
	in                   midi.In
}

func (s *Server) startHTTP() {
	// serve the website
	http.Handle("/", http.FileServer(http.Dir("./gui/dist")))
	//serve favicon
	http.Handle("/favicon.ico", http.FileServer(http.Dir("./gui/build/favicon.ico")))
	// serve /api
	http.Handle("/api/midi/", s.handleAPI())
	// handle websocket
	http.HandleFunc("/ws", s.wsEndpoint)
	// http listener
	log.Println("HTTP Listening on 8080")
	http.ListenAndServe(":8080", nil)
}

func (s *Server) Run() {

	ctx, cancel := context.WithCancel(context.Background())

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
	s.feedbackChannels = make(map[string]chan interface{})
	go s.sendNotes()

	s.websockets = make(map[string]*websocket.Conn)

	// record to a file
	if !s.DontRecord {
		s.in = common.GetMidiInput(drv, s.MidiPort)
		common.SetupCloseHandler(cancel, s.out)
		go recorder.Record(ctx, s.in)
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
		feedbackChan := make(chan interface{})
		key := c.RemoteAddr().String()
		s.feedbackChannelMutex.Lock()
		s.feedbackChannels[key] = feedbackChan
		s.feedbackChannelMutex.Unlock()

		enc := gob.NewEncoder(c)
		ctx2, cancel2 := context.WithCancel(context.Background())
		go func() {
			for {
				select {
				case feedback := <-feedbackChan:
					err := enc.Encode(types.TCPMessage{Body: feedback})
					common.Cont(err)
				case <-ctx2.Done():
					s.feedbackChannelMutex.Lock()
					delete(s.feedbackChannels, key)
					s.feedbackChannelMutex.Unlock()
					return
				}
			}
		}()
		go func() {
			dec := gob.NewDecoder(c)
			for {
				var t types.TCPMessage
				err := dec.Decode(&t)
				if err == io.EOF {
					log.Println("Connection closed by client.")
					cancel2()
					c.Close()
					return
				}
				common.Cont(err)
				// send through the channel
				s.notesChan <- t.Body
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

		// send it back through all feedback channels - if sending it wouldn't block
		for _, ch := range s.feedbackChannels {
			select {
			case ch <- input:
			default:
			}
		}

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
				go common.ExpandAllNotesOff(m, ms, s.MidiTuxChan, s.out)
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
