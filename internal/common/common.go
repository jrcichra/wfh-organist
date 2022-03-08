package common

import (
	"encoding/gob"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/jrcichra/wfh-organist/internal/types"
	"gitlab.com/gomidi/midi"
	driver "gitlab.com/gomidi/rtmididrv"
)

const LOW_VOLUME = 50
const HIGH_VOLUME = 100

func HandleMs(m time.Time) int64 {
	ms := time.Since(m).Milliseconds()
	return ms
}

func Must(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func Cont(err error) {
	if err != nil {
		log.Println(err)
	}
}

func PrintPort(port midi.Port) {
	log.Printf("[%v] %s\n", port.Number(), port.String())
}

func PrintOutPorts(ports []midi.Out) {
	log.Printf("MIDI OUT Ports\n")
	for _, port := range ports {
		PrintPort(port)
	}
	log.Printf("\n\n")
}

func PrintInPorts(ports []midi.In) {
	log.Printf("MIDI IN Ports\n")
	for _, port := range ports {
		PrintPort(port)
	}
	log.Printf("\n\n")
}

func GetLists() {
	drv, err := driver.New()
	Must(err)

	defer drv.Close()

	ins, err := drv.Ins()
	Must(err)

	outs, err := drv.Outs()
	Must(err)

	PrintInPorts(ins)
	PrintOutPorts(outs)
}

func ExpandAllNotesOff(m types.Raw, ms int64, midiTuxChan chan types.MidiTuxMessage, out midi.Out) {
	// for all channels
	var channel byte
	for channel = 0; channel < 16; channel++ {
		firstByte := channel + 0x90
		for k := uint8(0); k <= 0x7F; k++ {
			midiTuxChan <- types.MidiTuxMessage{
				Color: color.FgHiRed,
				T:     m,
				Ms:    ms,
			}
			// dont overwhelm the midi output
			time.Sleep(1 * time.Millisecond)
			_, err := out.Write([]byte{firstByte, k, 0})
			Cont(err)
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
			Cont(err)
		}
	}
}

func CheckAllNotesOff(data []byte) bool {
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

func SetupCloseHandler(out midi.Out, stopChan chan bool) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Println("\r- Ctrl+C pressed in Terminal. Turning off all notes and stopping the player/recorder.")
		select {
		case stopChan <- true:
		default:
		}
		expandAllNotesOffSignal(out)
		log.Println("Exiting...")
		os.Exit(0)
	}()
}

func RegisterGobTypes() {
	gob.Register(types.NoteOn{})
	gob.Register(types.NoteOff{})
	gob.Register(types.ProgramChange{})
	gob.Register(types.Aftertouch{})
	gob.Register(types.ControlChange{})
	gob.Register(types.NoteOffVelocity{})
	gob.Register(types.Pitchbend{})
	gob.Register(types.PolyAftertouch{})
	gob.Register(types.Raw{})
}