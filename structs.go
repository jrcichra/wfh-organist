package main

import "time"

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

// Get around gob types
type TCPMessage struct {
	Body interface{}
}
