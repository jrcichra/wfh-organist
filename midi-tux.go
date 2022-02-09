package main

import (
	"log"

	"github.com/fatih/color"
)

// Midi Tux prints colorful messages to the console
// Inspired by http://www.midiox.com/

func midiTux(midiTuxChan chan MidiTuxMessage) {
	for {
		midiTuxMsg := <-midiTuxChan
		midiTuxPrint(midiTuxMsg.Color, midiTuxMsg.T, midiTuxMsg.Ms)
	}
}

// this should only be called from the midiTux func
func midiTuxPrint(clr color.Attribute, t interface{}, ms int64) {
	color.Set(clr)
	switch m := t.(type) {
	case NoteOn:
		log.Printf("Type: %s, Channel: %2d, Key: %3d, Velocity: %2d, %d ms\n", "Note On ", m.Channel+1, m.Key, m.Velocity, ms)
	case NoteOff:
		log.Printf("Type: %s, Channel: %2d, Key: %3d, Velocity: %2d, %d ms\n", "Note Off", m.Channel+1, m.Key, 0, ms)
	case ProgramChange:
		log.Printf("Type: %s, Channel: %2d, Program: %2d, %d ms\n", "Program Change", m.Channel+1, m.Program, ms)
	case Aftertouch:
		log.Printf("Type: %s, Channel: %2d, Pressure: %2d, %d ms\n", "Aftertouch", m.Channel+1, m.Pressure, ms)
	case ControlChange:
		log.Printf("Type: %s, Channel: %2d, Controller: %2d, Value: %2d, %d ms\n", "Control Change", m.Channel+1, m.Controller, m.Value, ms)
	case NoteOffVelocity:
		log.Printf("Type: %s, Channel: %2d, Key: %3d, Velocity: %2d, %d ms\n", "Note Off Velocity", m.Channel+1, m.Key, m.Velocity, ms)
	case Pitchbend:
		log.Printf("Type: %s, Channel: %2d, Value: %3d, AbsValue: %4d, %d ms\n", "Pitchbend", m.Channel+1, m.Value, m.AbsValue, ms)
	case PolyAftertouch:
		log.Printf("Type: %s, Channel: %2d, Key: %3d, Pressure: %2d, %d ms\n", "Poly Aftertouch", m.Channel+1, m.Key, m.Pressure, ms)
	case Raw:
		log.Printf("Type: %s, Content: %x, %d ms\n", "Raw", m.Data, ms)
	default:
		log.Printf("Type: %s,\n", "Unknown")
	}
	color.Unset()
}
