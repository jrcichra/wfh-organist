package main

import "time"

// This is neccessary because NoteOn and NoteOff do not expose fields, so gob can't encode them.
// Since I'm only dealing with NoteOn and NoteOff, I can key off of the Velocity field.
type TCPMessage struct {
	Time     time.Time
	Channel  uint8
	Key      uint8
	Velocity uint8
}
