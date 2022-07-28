// play midi files

package player

import (
	"context"
	"log"
	"time"

	"github.com/jrcichra/wfh-organist/internal/types"
	"gitlab.com/gomidi/midi"
	"gitlab.com/gomidi/midi/midimessage/channel"
	"gitlab.com/gomidi/midi/player"
)

// https://pkg.go.dev/gitlab.com/gomidi/midi/player#Player
func PlayMidiFile(ctx context.Context, notesChan chan interface{}, file string, wrap bool) {
	log.Println("Playing midi file:", file)
	player, err := player.SMF(file)
	if err != nil {
		log.Println(err)
		return
	}
	player.GetMessages(func(wait time.Duration, m midi.Message, track int16) {
		select {
		case <-ctx.Done():
			return
		default:
		}
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
					Time:     time.Now(),
				}
			} else {
				notesChan <- v
			}
		case channel.NoteOff:
			if wrap {
				notesChan <- types.NoteOff{
					Channel: v.Channel(),
					Key:     v.Key(),
					Time:    time.Now(),
				}
			} else {
				notesChan <- v
			}
		case channel.ProgramChange:
			if wrap {
				notesChan <- types.ProgramChange{
					Channel: v.Channel(),
					Program: v.Program(),
					Time:    time.Now(),
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
					Time:       time.Now(),
				}
			} else {
				notesChan <- v
			}
		}
	})
	<-ctx.Done()
}
