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
func PlayMidiFile(notesChan chan interface{}, file string, stop chan struct{}) {

	log.Println("Playing midi file:", file)
	stopPlayChan := make(chan struct{})
	stopBool := false
	player, err := player.SMF(file)
	if err != nil {
		log.Println(err)
		return
	}
	player.GetMessages(func(wait time.Duration, m midi.Message, track int16) {

		// check if we should stop
		select {
		case <-stop:
			stopPlayChan <- struct{}{} // this frees up the player resources
			// these GetMessages might still be around so we need to update a bool to not sound them and finish the file
			stopBool = true
		default:
		}
		if !stopBool {
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
		}
	})
	// sleep until asked to stop
	<-stopPlayChan
}
