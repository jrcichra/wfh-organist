package main

import (
	"encoding/gob"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fatih/color"
	"gitlab.com/gomidi/midi"
	driver "gitlab.com/gomidi/rtmididrv"
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

func expandAllNotesOff(m Raw, ms int64, midiTuxChan chan MidiTuxMessage, out midi.Out) {
	// for all channels
	var channel byte
	for channel = 0; channel < 16; channel++ {
		firstByte := channel + 0x90
		for k := uint8(0); k <= 0x7F; k++ {
			midiTuxChan <- MidiTuxMessage{
				Color: color.FgHiRed,
				T:     m,
				Ms:    ms,
			}
			// dont overwhelm the midi output
			time.Sleep(1 * time.Millisecond)
			_, err := out.Write([]byte{firstByte, k, 0})
			cont(err)
		}
	}
}

func expandAllNotesOffSignal(out midi.Out) {
	// for all channels
	var channel byte
	for channel = 0; channel < 16; channel++ {
		firstByte := channel + 0x90
		for k := uint8(0); k <= 0x7F; k++ {
			// dont overwhelm the midi output
			time.Sleep(1 * time.Millisecond)
			_, err := out.Write([]byte{firstByte, k, 0})
			cont(err)
		}
	}
}

func checkAllNotesOff(data []byte) bool {
	firstByte := data[0]
	secondByte := data[1]
	thirdByte := data[2]
	switch firstByte {
	case 0xb0, 0xb1, 0xb2, 0xb3, 0xb4, 0xb5, 0xb6, 0xb7, 0xb8, 0xb9, 0xba, 0xbb, 0xbc, 0xbd, 0xbe, 0xbf:
		if secondByte == 0x7b && thirdByte == 0x00 {
			return true
		} else {
			return false
		}
	default:
		return false
	}
}

func SetupCloseHandler(out midi.Out) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Println("\r- Ctrl+C pressed in Terminal. Turning off all notes.")
		expandAllNotesOffSignal(out)
		log.Println("Exiting...")
		os.Exit(0)
	}()
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
	gob.Register(Raw{})
}
