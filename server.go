package main

import (
	"encoding/gob"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/fatih/color"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gitlab.com/gomidi/midi/writer"
	driver "gitlab.com/gomidi/rtmididrv"
)

func handleMs(m time.Time) int64 {
	ms := time.Since(m).Milliseconds()
	prom_last_ms.Set(float64(ms))
	return ms
}

func startHTTP(notesChan chan interface{}) {
	// serve metrics
	http.Handle("/metrics", promhttp.Handler())
	// serve the website
	// http.Handle("/", http.FileServer(http.Dir("./static")))
	// serve /api
	http.Handle("/api/midi/raw", handleAPI(notesChan))
	// http listener
	log.Println("HTTP Listening on 8080")
	http.ListenAndServe(":8080", nil)
}

func server(midiPort int, serverPort int, protocol string) {

	// wait for someone to connect to the server
	l, err := net.Listen(protocol, ":"+strconv.Itoa(serverPort))
	must(err)
	defer l.Close()

	//send notes listening to a go channel
	notesChan := make(chan interface{})
	go sendNotes(midiPort, notesChan)
	// also can accept notes from the HTTP API
	go startHTTP(notesChan)

	// keep accepting connections
	for {
		log.Println("Listening on", l.Addr())
		c, err := l.Accept()
		must(err)
		log.Println("Connection from:", c.RemoteAddr())
		log.Println("Ready to play music!")

		go func() {
			for {
				dec := gob.NewDecoder(c)
				var t TCPMessage
				err := dec.Decode(&t)
				if err == io.EOF {
					log.Println("Connection closed by client.")
					c.Close()
					return
				}
				must(err)
				// send through the channel
				notesChan <- t.Body
			}
		}()
	}
}

func sendNotes(midiPort int, notesChan chan interface{}) {

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
			midiTuxServerPrint(color.FgHiGreen, m, ms)
		case NoteOff:
			ms := handleMs(m.Time)
			cont(writer.NoteOff(writers[m.Channel], m.Key))
			midiTuxServerPrint(color.FgHiRed, m, ms)
		case ProgramChange:
			ms := handleMs(m.Time)
			cont(writer.ProgramChange(writers[m.Channel], m.Program))
			midiTuxServerPrint(color.FgHiYellow, m, ms)
		case Aftertouch:
			ms := handleMs(m.Time)
			cont(writer.Aftertouch(writers[m.Channel], m.Pressure))
			midiTuxServerPrint(color.FgHiBlue, m, ms)
		case ControlChange:
			ms := handleMs(m.Time)
			cont(writer.ControlChange(writers[m.Channel], m.Controller, m.Value))
			midiTuxServerPrint(color.FgHiMagenta, m, ms)
		case NoteOffVelocity:
			ms := handleMs(m.Time)
			cont(writer.NoteOffVelocity(writers[m.Channel], m.Key, m.Velocity))
			midiTuxServerPrint(color.FgHiYellow, m, ms)
		case Pitchbend:
			ms := handleMs(m.Time)
			cont(writer.Pitchbend(writers[m.Channel], m.Value))
			midiTuxServerPrint(color.FgMagenta, m, ms)
		case PolyAftertouch:
			ms := handleMs(m.Time)
			cont(writer.PolyAftertouch(writers[m.Channel], m.Key, m.Pressure))
			midiTuxServerPrint(color.FgCyan, m, ms)
		case Raw:
			ms := handleMs(m.Time)
			midiTuxServerPrint(color.FgBlue, m, ms)
			// write the raw bytes to the MIDI device
			_, err := out.Write(m.Data)
			cont(err)
		default:
			log.Println("Unknown message type:", m)
		}
	}
}
