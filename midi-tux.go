package main

import (
	"log"
	"net"

	"github.com/fatih/color"
)

// Midi Tux prints colorful messages to the console
// Inspired by http://www.midiox.com/

func midiTuxPrint(clr color.Attribute, addr net.Addr, t interface{}, ms int64) {
	color.Set(clr)
	switch m := t.(type) {
	case NoteOn:
		log.Printf("User: %s, Type: %s, Channel: %2d, Key: %3d, Velocity: %2d, Delay: %4d ms\n", addr, "Note On", m.Channel, m.Key, m.Velocity, ms)
	case NoteOff:
		log.Printf("User: %s, Type: %s, Channel: %2d, Key: %3d, Velocity: %2d, Delay: %4d ms\n", addr, "Note Off", m.Channel, m.Key, 0, ms)
	case ProgramChange:
		log.Printf("User: %s, Type: %s, Channel: %2d, Program: %2d, Delay: %4d ms\n", addr, "Program Change", m.Channel, m.Program, ms)
	case Aftertouch:
		log.Printf("User: %s, Type: %s, Channel: %2d, Pressure: %2d, Delay: %4d ms\n", addr, "Aftertouch", m.Channel, m.Pressure, ms)
	case ControlChange:
		log.Printf("User: %s, Type: %s, Channel: %2d, Controller: %2d, Value: %2d, Delay: %4d ms\n", addr, "Control Change", m.Channel, m.Controller, m.Value, ms)
	case NoteOffVelocity:
		log.Printf("User: %s, Type: %s, Channel: %2d, Key: %3d, Velocity: %2d, Delay: %4d ms\n", addr, "Note Off Velocity", m.Channel, m.Key, m.Velocity, ms)
	case Pitchbend:
		log.Printf("User: %s, Type: %s, Channel: %2d, Value: %3d, AbsValue: %4d, Delay: %4d ms\n", addr, "Pitchbend", m.Channel, m.Value, m.AbsValue, ms)
	case PolyAftertouch:
		log.Printf("User: %s, Type: %s, Channel: %2d, Key: %3d, Pressure: %2d, Delay: %4d ms\n", addr, "Poly Aftertouch", m.Channel, m.Key, m.Pressure, ms)
	case Raw:
		log.Printf("User: %s, Type: %s, Content: %s Delay: %4d ms\n", addr, "Raw", string(m.Data), ms)
	default:
		log.Printf("User: %s, Type: %s, Delay: %4d ms\n", addr, "Unknown", ms)
	}
	color.Unset()
}
