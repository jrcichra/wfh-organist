package main

import (
	"flag"
	"log"
	_ "net/http/pprof"
	"os"
	"strings"

	"github.com/jrcichra/wfh-organist/internal/client"
	"github.com/jrcichra/wfh-organist/internal/common"
	"github.com/jrcichra/wfh-organist/internal/miditux"
	"github.com/jrcichra/wfh-organist/internal/server"
	"github.com/jrcichra/wfh-organist/internal/types"
)

func main() {
	log.SetOutput(os.Stdout)

	// get args
	serverIP := flag.String("server", "localhost", "server IP")
	serverPort := flag.Int("port", 3131, "server port")
	midiPort := flag.Int("midi", 1, "midi port")
	list := flag.Bool("list", false, "list available ports")
	mode := flag.String("mode", "local", "client, server, or local (runs both)")
	protocol := flag.String("protocol", "tcp", "tcp only (udp not implemented yet)")
	profile := flag.String("profile", "profiles/wosp/", "profiles path")
	stdinMode := flag.Bool("stdin", false, "read from stdin")
	delay := flag.Int("delay", 0, "artificial delay in ms")
	file := flag.String("file", "", "midi file to play")
	dontControlVolume := flag.Bool("novolume", false, "have WFHO control client volume")
	dontRecord := flag.Bool("norecord", false, "continuously record midi")

	flag.Parse()

	// make sure stdinMode is only true if mode is client
	if *stdinMode && strings.ToLower(*mode) != "client" {
		log.Println("stdin mode can only be used with client mode")
		return
	}

	// delay only works on the client or local mode
	if *delay > 0 && strings.ToLower(*mode) != "client" && strings.ToLower(*mode) != "local" {
		log.Println("delay only works with client or local mode")
		return
	}

	// print MIDI IO if requested
	if *list {
		common.GetLists()
		return
	}

	switch *protocol {
	case "tcp":
		// do nothing
	default:
		log.Println("Invalid protocol", *protocol)
		return
	}

	// register types to gob
	common.RegisterGobTypes()

	// spin up a midi-tux goroutine to handle message outputs
	midiTuxChan := make(chan types.MidiTuxMessage, 100)
	go miditux.MidiTux(midiTuxChan)

	// operate in client or server mode
	switch strings.ToLower(*mode) {
	case "server":
		go server.Server(*midiPort, *serverPort, *protocol, midiTuxChan, *dontRecord, *profile)
	case "client":
		go client.Client(*midiPort, *serverIP, *serverPort, *protocol, *stdinMode, *delay, *file, midiTuxChan, *profile, *dontControlVolume)
	case "local":
		// run both (unless serverIP is set, and sleep forever
		if *serverIP == "localhost" {
			go server.Server(*midiPort, *serverPort, *protocol, midiTuxChan, *dontRecord, *profile)
		}
		go client.Client(*midiPort, *serverIP, *serverPort, *protocol, *stdinMode, *delay, *file, midiTuxChan, *profile, *dontControlVolume)
	default:
		log.Fatalf("Unknown mode: %s. Must be 'server' or 'client'\n", *mode)
	}
	select {}
}
