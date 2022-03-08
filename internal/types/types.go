package types

import (
	"time"

	"github.com/fatih/color"
)

type MidiCSVRecord struct {
	InputChannel  uint8
	OutputChannel uint8
	Offset        int
}

type NoteOn struct {
	Time     time.Time
	Channel  uint8
	Key      uint8
	Velocity uint8
}

type NoteOff struct {
	Time    time.Time
	Channel uint8
	Key     uint8
}

type ProgramChange struct {
	Time    time.Time
	Channel uint8
	Program uint8
}

type Aftertouch struct {
	Time     time.Time
	Channel  uint8
	Pressure uint8
}

type ControlChange struct {
	Time       time.Time
	Channel    uint8
	Controller uint8
	Value      uint8
}

type NoteOffVelocity struct {
	Time     time.Time
	Channel  uint8
	Key      uint8
	Velocity uint8
}

type Pitchbend struct {
	Time     time.Time
	Channel  uint8
	Value    int16
	AbsValue uint16
}

type PolyAftertouch struct {
	Time     time.Time
	Channel  uint8
	Key      uint8
	Pressure uint8
}

// Raw sends raw bytes to the server
type Raw struct {
	Time time.Time
	Data []byte
}

// Info sends information in a single key/value format - used for bootstrapping services like trx
type Info struct {
	Key   string
	Value string
}

// Get around gob types
type TCPMessage struct {
	Body interface{}
}

//Messages sent to MidiTux to print
type MidiTuxMessage struct {
	Color color.Attribute
	T     interface{}
	Ms    int64
}
