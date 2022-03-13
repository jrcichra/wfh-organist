package recorder

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"sync"
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

func Record(in midi.In, stop chan bool) {
	resolution := smf.MetricTicks(1920)
	bpm := 120.00
	external := false

	for !external {

		timer := timer.Timer{}
		timeout := timer.New(1 * 60)

		var inbf bytes.Buffer
		var outbf bytes.Buffer

		waitForFirstNote := true

		var wr *writer.SMF
		rd := midireader.New(&inbf, nil)
		ch := make(chan timedMsg)
		internalStop := make(chan struct{})

		var wg sync.WaitGroup
		wg.Add(1)

		go func() {
			for {
				select {
				case tm := <-ch:
					deltaticks := resolution.FractionalTicks(bpm, time.Duration(tm.deltaMicrosecs)*time.Microsecond)
					wr.SetDelta(deltaticks)
					inbf.Write(tm.data)
					msg, _ := rd.Read()
					wr.Write(msg)
				case <-internalStop:
					wg.Done()
					return
				case <-timeout:
					log.Println("Recording has Timed out")
					stop <- false // force a stop, no return or waitgroup so we hit internalStop next loop
				}
			}
		}()

		in.SetListener(func(data []byte, deltaMicrosecs int64) {
			if len(data) == 0 {
				return
			}

			// Probably introduces a race condition
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

		// wait until we're told to stop (by the outside world or ourself)
		external = <-stop
		in.StopListening()
		internalStop <- struct{}{}
		wg.Wait()

		if wr != nil {
			wr.Write(meta.EndOfTrack)
			// get the epoch
			file := fmt.Sprintf("recordings/%d.mid", time.Now().Unix())
			log.Println("Writing to", file)
			ioutil.WriteFile(file, outbf.Bytes(), 0644)
		}
	}

}
