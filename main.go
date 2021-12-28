package main

import (
	"flag"
	"log"
	"strconv"
	"strings"
)

func main() {

	// get args
	serverIP := flag.String("server", "localhost", "server IP")
	serverPort := flag.Int("port", 3131, "server port")
	defaultMidiPort := 0 // flag does not print default value on 0 int
	midiPort := flag.Int("midi", defaultMidiPort, "midi port (default "+strconv.Itoa(defaultMidiPort)+")")
	list := flag.Bool("list", false, "list available ports")
	mode := flag.String("mode", "local", "client, server, or local (runs both)")
	flag.Parse()

	// print MIDI IO if requested
	if *list {
		getLists()
		return
	}

	// operate in client or server mode
	switch strings.ToLower(*mode) {
	case "server":
		server(*midiPort, *serverPort)
	case "client":
		client(*midiPort, *serverIP, *serverPort)
	case "local":
		// run both and sleep forever
		go server(*midiPort, *serverPort)
		go client(*midiPort, *serverIP, *serverPort)
		select {}
	default:
		log.Printf("Unknown mode: %s. Must be 'server' or 'client'\n", *mode)
	}
}
