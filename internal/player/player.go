// play midi files

package player

import (
	"log"
	"time"

	"gitlab.com/gomidi/midi"
	"gitlab.com/gomidi/midi/midimessage/channel"
	"gitlab.com/gomidi/midi/player"
)

// https://pkg.go.dev/gitlab.com/gomidi/midi/player#Player
func PlayMidiFile(notesChan chan interface{}, file string) {

	log.Println("Playing midi file:", file)

	player, err := player.SMF(file)
	if err != nil {
		log.Fatal(err)
	}
	player.GetMessages(func(wait time.Duration, m midi.Message, track int16) {
		// sleep for the wait amount
		time.Sleep(wait)
		// send the message to the channel if it's a noteon or noteoff
		switch v := m.(type) {
		case channel.NoteOn:
			notesChan <- v
		case channel.NoteOff:
			notesChan <- v
		case channel.ProgramChange:
			notesChan <- v
		case channel.ControlChange:
			notesChan <- v
		}
	})
	// sleep forever
	select {}
}
