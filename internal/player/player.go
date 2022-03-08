// play midi files

package player

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jrcichra/wfh-organist/internal/common"
	"github.com/jrcichra/wfh-organist/internal/types"
	"gitlab.com/gomidi/midi"
	"gitlab.com/gomidi/midi/midimessage/channel"
	"gitlab.com/gomidi/midi/player"
)

// https://pkg.go.dev/gitlab.com/gomidi/midi/player#Player
func PlayMidiFile(notesChan chan interface{}, file string, stopPlayingChan chan bool, wrap bool) {

	log.Println("Playing midi file:", file)
	stopRoutine := make(chan struct{})
	stopBool := false
	player, err := player.SMF(file)
	if err != nil {
		log.Println(err)
		return
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		common.ShutdownWg.Add(1) // only add to the waitgroup once we got the signal, there might be multiple players and I'm not ready to handle that case yet
		log.Println("\r- Ctrl+C pressed in Terminal. Stopping playback...")
		<-stopPlayingChan
	}()

	go func() {
		// wait for when we should stop
		<-stopPlayingChan
		// these GetMessages might still be around so we need to update a bool to not sound them and finish the file
		stopBool = true
		stopRoutine <- struct{}{}
		common.ShutdownWg.Done()
	}()

	player.GetMessages(func(wait time.Duration, m midi.Message, track int16) {

		if !stopBool {
			// sleep for the wait amount
			time.Sleep(wait)
			// send the message to the channel if it's a noteon or noteoff
			switch v := m.(type) {
			case channel.NoteOn:
				if wrap {
					notesChan <- types.NoteOn{
						Channel:  v.Channel(),
						Key:      v.Key(),
						Velocity: v.Velocity(),
					}
				} else {
					notesChan <- v
				}
			case channel.NoteOff:
				if wrap {
					notesChan <- types.NoteOff{
						Channel: v.Channel(),
						Key:     v.Key(),
					}
				} else {
					notesChan <- v
				}
			case channel.ProgramChange:
				if wrap {
					notesChan <- types.ProgramChange{
						Channel: v.Channel(),
						Program: v.Program(),
					}
				} else {
					notesChan <- v
				}
			case channel.ControlChange:
				if wrap {
					notesChan <- types.ControlChange{
						Channel:    v.Channel(),
						Controller: v.Controller(),
						Value:      v.Value(),
					}
				} else {
					notesChan <- v
				}
			}
		}
	})
	// sleep until asked to stop
	<-stopRoutine
}
