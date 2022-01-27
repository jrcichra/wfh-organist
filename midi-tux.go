package main

import (
	"log"

	"github.com/fatih/color"
	"gitlab.com/gomidi/midi/midimessage/channel"
)

// Midi Tux prints colorful messages to the console
// Inspired by http://www.midiox.com/

func midiTuxServerPrint(clr color.Attribute, t interface{}, ms int64) {
	color.Set(clr)
	switch m := t.(type) {
	case NoteOn:
		log.Printf("Type: %s, Channel: %2d, Key: %3d, Velocity: %2d,\n", "Note On", m.Channel, m.Key, m.Velocity)
	case NoteOff:
		log.Printf("Type: %s, Channel: %2d, Key: %3d, Velocity: %2d,\n", "Note Off", m.Channel, m.Key, 0)
	case ProgramChange:
		log.Printf("Type: %s, Channel: %2d, Program: %2d,\n", "Program Change", m.Channel, m.Program)
	case Aftertouch:
		log.Printf("Type: %s, Channel: %2d, Pressure: %2d,\n", "Aftertouch", m.Channel, m.Pressure)
	case ControlChange:
		log.Printf("Type: %s, Channel: %2d, Controller: %2d, Value: %2d,\n", "Control Change", m.Channel, m.Controller, m.Value)
	case NoteOffVelocity:
		log.Printf("Type: %s, Channel: %2d, Key: %3d, Velocity: %2d,\n", "Note Off Velocity", m.Channel, m.Key, m.Velocity)
	case Pitchbend:
		log.Printf("Type: %s, Channel: %2d, Value: %3d, AbsValue: %4d,\n", "Pitchbend", m.Channel, m.Value, m.AbsValue)
	case PolyAftertouch:
		log.Printf("Type: %s, Channel: %2d, Key: %3d, Pressure: %2d,\n", "Poly Aftertouch", m.Channel, m.Key, m.Pressure)
	case Raw:
		log.Printf("Type: %s, Content: %x\n", "Raw", m.Data)
	default:
		log.Printf("Type: %s,\n", "Unknown")
	}
	color.Unset()
}

func midiTuxClientPrint(clr color.Attribute, t interface{}, newChannel uint8, newKey uint8) {
	color.Set(clr)
	switch m := t.(type) {
	case channel.NoteOn:
		log.Printf("Type: %s, Old Channel: %2d,  New Channel: %2d, Old Key %3d, New Key: %3d, Velocity: %2d\n", "Note On", m.Channel(), newChannel, m.Key(), newKey, m.Velocity())
	case channel.NoteOff:
		log.Printf("Type: %s, Old Channel: %2d,  New Channel: %2d, Old Key %3d, New Key: %3d\n", "Note Off", m.Channel(), newChannel, m.Key(), newKey)
	case channel.ProgramChange:
		log.Printf("Type: %s, Old Channel: %2d,  New Channel: %2d, Program: %2d,\n", "Program Change", m.Channel(), newChannel, m.Program())
	case channel.Aftertouch:
		log.Printf("Type: %s, Old Channel: %2d,  New Channel: %2d, Pressure: %2d,\n", "Aftertouch", m.Channel(), newChannel, m.Pressure())
	case channel.ControlChange:
		log.Printf("Type: %s, Old Channel: %2d,  New Channel: %2d, Controller: %2d, Value: %2d,\n", "Control Change", m.Channel(), newChannel, m.Controller(), m.Value())
	case channel.NoteOffVelocity:
		log.Printf("Type: %s, Old Channel: %2d,  New Channel: %2d, Old Key: %3d, New Key: %3d, Velocity: %2d,\n", "Note Off Velocity", m.Channel(), newChannel, m.Key(), newKey, m.Velocity())
	case channel.Pitchbend:
		log.Printf("Type: %s, Old Channel: %2d,  New Channel: %2d, Value: %3d, AbsValue: %4d,\n", "Pitchbend", m.Channel(), newChannel, m.Value(), m.AbsValue())
	case channel.PolyAftertouch:
		log.Printf("Type: %s, Old Channel: %2d,  New Channel: %2d, Old Key: %3d, New Key: %3d, Pressure: %2d,\n", "Poly Aftertouch", m.Channel(), newChannel, m.Key(), newKey, m.Pressure())
	case Raw:
		log.Printf("Type: %s, Content: %x\n", "Raw", m.Data)
	default:
		log.Printf("Type: %s\n", "Unknown")
	}
	color.Unset()
}
