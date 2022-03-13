package miditux

import (
	"log"
	"strings"

	"github.com/fatih/color"
	"github.com/jrcichra/wfh-organist/internal/types"
)

// Midi Tux prints colorful messages to the console
// Inspired by http://www.midiox.com/

func MidiTux(midiTuxChan chan types.MidiTuxMessage) {
	for {
		midiTuxMsg := <-midiTuxChan
		midiTuxPrint(midiTuxMsg.Color, midiTuxMsg.T, midiTuxMsg.Ms)
	}
}

// this should only be called from the midiTux func
func midiTuxPrint(clr color.Attribute, t interface{}, ms int64) {
	slowStr := strings.Repeat("*", int(ms/100))
	color.Set(clr)
	switch m := t.(type) {
	case types.NoteOn:
		log.Printf("Type: %s, Channel: %2d, Key: %3d, Velocity: %2d, %d ms %s\n", "Note On ", m.Channel+1, m.Key, m.Velocity, ms, slowStr)
	case types.NoteOff:
		log.Printf("Type: %s, Channel: %2d, Key: %3d, Velocity: %2d, %d ms %s\n", "Note Off", m.Channel+1, m.Key, 0, ms, slowStr)
	case types.ProgramChange:
		log.Printf("Type: %s, Channel: %2d, Program: %2d, %d ms%s\n", "Program Change", m.Channel+1, m.Program, ms, slowStr)
	case types.Aftertouch:
		log.Printf("Type: %s, Channel: %2d, Pressure: %2d, %d ms%s\n", "Aftertouch", m.Channel+1, m.Pressure, ms, slowStr)
	case types.ControlChange:
		log.Printf("Type: %s, Channel: %2d, Controller: %2d, Value: %2d, %d ms %s\n", "Control Change", m.Channel+1, m.Controller, m.Value, ms, slowStr)
	case types.NoteOffVelocity:
		log.Printf("Type: %s, Channel: %2d, Key: %3d, Velocity: %2d, %d ms %s\n", "Note Off Velocity", m.Channel+1, m.Key, m.Velocity, ms, slowStr)
	case types.Pitchbend:
		log.Printf("Type: %s, Channel: %2d, Value: %3d, AbsValue: %4d, %d ms %s\n", "Pitchbend", m.Channel+1, m.Value, m.AbsValue, ms, slowStr)
	case types.PolyAftertouch:
		log.Printf("Type: %s, Channel: %2d, Key: %3d, Pressure: %2d, %d ms %s\n", "Poly Aftertouch", m.Channel+1, m.Key, m.Pressure, ms, slowStr)
	case types.Raw:
		log.Printf("Type: %s, Content: %x, %d ms %s\n", "Raw", m.Data, ms, slowStr)
	default:
		log.Printf("Type: %s,\n", "Unknown")
	}
	color.Unset()
}
