package main

import (
	"log"
	"sync"

	"github.com/fatih/color"
)

// Midi Tux prints colorful messages to the console
// Inspired by http://www.midiox.com/

var midiTuxColorMutex = &sync.Mutex{}

func midiTuxPrint(clr color.Attribute, t interface{}, ms int64) {
	midiTuxColorMutex.Lock()
	color.Set(clr)
	switch m := t.(type) {
	case NoteOn:
		log.Printf("Type: %s, Channel: %2d, Key: %3d, Velocity: %2d,\n", "Note On", m.Channel+1, m.Key, m.Velocity)
	case NoteOff:
		log.Printf("Type: %s, Channel: %2d, Key: %3d, Velocity: %2d,\n", "Note Off", m.Channel+1, m.Key, 0)
	case ProgramChange:
		log.Printf("Type: %s, Channel: %2d, Program: %2d,\n", "Program Change", m.Channel+1, m.Program)
	case Aftertouch:
		log.Printf("Type: %s, Channel: %2d, Pressure: %2d,\n", "Aftertouch", m.Channel+1, m.Pressure)
	case ControlChange:
		log.Printf("Type: %s, Channel: %2d, Controller: %2d, Value: %2d,\n", "Control Change", m.Channel+1, m.Controller, m.Value)
	case NoteOffVelocity:
		log.Printf("Type: %s, Channel: %2d, Key: %3d, Velocity: %2d,\n", "Note Off Velocity", m.Channel+1, m.Key, m.Velocity)
	case Pitchbend:
		log.Printf("Type: %s, Channel: %2d, Value: %3d, AbsValue: %4d,\n", "Pitchbend", m.Channel+1, m.Value, m.AbsValue)
	case PolyAftertouch:
		log.Printf("Type: %s, Channel: %2d, Key: %3d, Pressure: %2d,\n", "Poly Aftertouch", m.Channel+1, m.Key, m.Pressure)
	case Raw:
		log.Printf("Type: %s, Content: %x\n", "Raw", m.Data)
	default:
		log.Printf("Type: %s,\n", "Unknown")
	}
	color.Unset()
	midiTuxColorMutex.Unlock()
}
