package recorder

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/jrcichra/wfh-organist/pkg/timer"
	"gitlab.com/gomidi/midi"
	"gitlab.com/gomidi/midi/midimessage/meta"
	"gitlab.com/gomidi/midi/midimessage/meta/meter"
	"gitlab.com/gomidi/midi/midireader"
	"gitlab.com/gomidi/midi/smf"
	"gitlab.com/gomidi/midi/smf/smfwriter"
	"gitlab.com/gomidi/midi/writer"
)

// https://gitlab.com/gomidi/midi/-/blob/master/examples/liverecord/main.go

type timedMsg struct {
	deltaMicrosecs int64
	data           []byte
}

func Record(ctx context.Context, in midi.In) error {
	resolution := smf.MetricTicks(1920)
	bpm := 120.00

	timer := timer.NewTimer(10 * time.Second)

	var inbf bytes.Buffer
	var outbf bytes.Buffer

	waitForFirstNote := true

	var wr *writer.SMF
	rd := midireader.New(&inbf, nil)
	ch := make(chan timedMsg)

	go func() {
		for {
			select {
			case <-ctx.Done():
				// this thread should end because the program is shutting down
				return
			case tm := <-ch:
				deltaticks := resolution.FractionalTicks(bpm, time.Duration(tm.deltaMicrosecs)*time.Microsecond)
				wr.SetDelta(deltaticks)
				inbf.Write(tm.data)
				msg, _ := rd.Read()
				wr.Write(msg)
			case <-timer.Done():
				log.Println("Recording has timed out")
				if wr != nil {
					wr.Write(meta.EndOfTrack)
					// get the epoch
					file := fmt.Sprintf("recordings/%d.mid", time.Now().Unix())
					log.Println("Writing to", file)
					ioutil.WriteFile(file, outbf.Bytes(), 0644)
					// reset for a new file when a new note comes in
					waitForFirstNote = true
					// reset the buffers
					inbf.Reset()
					outbf.Reset()
				}
				timer.Reset()
			}
		}
	}()

	in.SetListener(func(data []byte, deltaMicrosecs int64) {
		if len(data) == 0 {
			return
		}

		// Probably racey but this is all I can come up with at the moment
		if waitForFirstNote {
			waitForFirstNote = false
			// start to write the file
			log.Println("Recording...")
			timer.Start()
			wr = writer.NewSMF(&outbf, 1, smfwriter.TimeFormat(resolution))
			wr.WriteHeader()
			wr.Write(meta.FractionalBPM(bpm)) // set the initial bpm
			wr.Write(meter.M4_4())            // set the meter if needed
		}

		// reset the timer
		timer.Reset()

		ch <- timedMsg{data: data, deltaMicrosecs: deltaMicrosecs}
	})

	// wait here until the program is shutting down
	<-ctx.Done()
	in.StopListening()
	log.Println("Recording routine ended")
	return ctx.Err()
}
