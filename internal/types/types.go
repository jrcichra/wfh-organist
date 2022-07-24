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

// func (m *MidiMap) init() {
// 	m.Maps = make(map[uint8]uint8)
// }

// func (m *MidiMap) add(input uint8, output uint8) {
// 	m.Maps[input] = output
// }

type NoteOn struct {
	Time     time.Time `json:"time"`
	Channel  uint8     `json:"channel"`
	Key      uint8     `json:"key"`
	Velocity uint8     `json:"velocity"`
}

type NoteOff struct {
	Time    time.Time `json:"time"`
	Channel uint8     `json:"channel"`
	Key     uint8     `json:"key"`
}

type ProgramChange struct {
	Time    time.Time `json:"time"`
	Channel uint8     `json:"channel"`
	Program uint8     `json:"program"`
}

type Aftertouch struct {
	Time     time.Time `json:"time"`
	Channel  uint8     `json:"channel"`
	Pressure uint8     `json:"pressure"`
}

type ControlChange struct {
	Time       time.Time `json:"time"`
	Channel    uint8     `json:"channel"`
	Controller uint8     `json:"controller"`
	Value      uint8     `json:"value"`
}

type NoteOffVelocity struct {
	Time     time.Time `json:"time"`
	Channel  uint8     `json:"channel"`
	Key      uint8     `json:"key"`
	Velocity uint8     `json:"velocity"`
}

type Pitchbend struct {
	Time     time.Time `json:"time"`
	Channel  uint8     `json:"channel"`
	Value    int16     `json:"value"`
	AbsValue uint16    `json:"absvalue"`
}

type PolyAftertouch struct {
	Time     time.Time `json:"time"`
	Channel  uint8     `json:"channel"`
	Key      uint8     `json:"key"`
	Pressure uint8     `json:"pressure"`
}

// Raw sends raw bytes to the server
type Raw struct {
	Time time.Time `json:"time"`
	Data []byte    `json:"data"`
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
