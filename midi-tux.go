package main

import (
	"log"
	"net"

	"github.com/fatih/color"
)

// Midi Tux prints colorful messages to the console
// Inspired by http://www.midiox.com/

func midiTuxPrint(clr color.Attribute, addr net.Addr, t TCPMessage, ms int64) {
	var typ string
	if t.Velocity == 0 {
		typ = "Note Off"
	} else {
		typ = "Note On"
	}
	color.Set(clr)
	log.Printf("User: %s, Type: %s, Channel: %2d, Key: %3d, Velocity: %2d, Delay: %4d ms\n", addr, typ, t.Channel, t.Key, t.Velocity, ms)
	color.Unset()
}
