package main

import (
	"encoding/gob"
	"log"

	"gitlab.com/gomidi/midi"
	driver "gitlab.com/gomidi/rtmididrv"
)

const (
	NOTEON          = "Note On"
	NOTEOFF         = "Note Off"
	PROGRAMCHANGE   = "Program Change"
	AFTERTOUCH      = "Aftertouch"
	CONTROLCHANGE   = "Control Change"
	NOTEOFFVELOCITY = "Note Off Velocity"
	PITCHBEND       = "Pitchbend"
	POLYAFTERTOUCH  = "Poly Aftertouch"
	// SYSTEMEXCLUSIVE="SYSTEMEXCLUSIVE"
)

func must(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func cont(err error) {
	if err != nil {
		log.Println(err)
	}
}

func printPort(port midi.Port) {
	log.Printf("[%v] %s\n", port.Number(), port.String())
}

func printOutPorts(ports []midi.Out) {
	log.Printf("MIDI OUT Ports\n")
	for _, port := range ports {
		printPort(port)
	}
	log.Printf("\n\n")
}

func printInPorts(ports []midi.In) {
	log.Printf("MIDI IN Ports\n")
	for _, port := range ports {
		printPort(port)
	}
	log.Printf("\n\n")
}

func getLists() {
	drv, err := driver.New()
	must(err)

	defer drv.Close()

	ins, err := drv.Ins()
	must(err)

	outs, err := drv.Outs()
	must(err)

	printInPorts(ins)
	printOutPorts(outs)
}

func registerGobTypes() {
	gob.Register(NoteOn{})
	gob.Register(NoteOff{})
	gob.Register(ProgramChange{})
	gob.Register(Aftertouch{})
	gob.Register(ControlChange{})
	gob.Register(NoteOffVelocity{})
	gob.Register(Pitchbend{})
	gob.Register(PolyAftertouch{})
}
