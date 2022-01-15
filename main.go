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
	protocol := flag.String("protocol", "tcp", "tcp only (udp not implemented yet)")
	stdinMode := flag.Bool("stdin", false, "read from stdin")
	flag.Parse()

	// make sure stdinMode is only true if mode is client
	if *stdinMode && strings.ToLower(*mode) != "client" {
		log.Println("stdin mode can only be used with client mode")
		return
	}

	// print MIDI IO if requested
	if *list {
		getLists()
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
	registerGobTypes()

	// operate in client or server mode
	switch strings.ToLower(*mode) {
	case "server":
		go server(*midiPort, *serverPort, *protocol)
	case "client":
		go client(*midiPort, *serverIP, *serverPort, *protocol, *stdinMode)
	case "local":
		// run both and sleep forever
		go server(*midiPort, *serverPort, *protocol)
		go client(*midiPort, *serverIP, *serverPort, *protocol, *stdinMode)
	default:
		log.Fatalf("Unknown mode: %s. Must be 'server' or 'client'\n", *mode)
	}
	select {}
}
