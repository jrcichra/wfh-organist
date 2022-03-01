package recorder

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"sync"
	"time"

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

func Record(in midi.In, stop chan struct{}) {
	log.Println("Recording...")
	var inbf bytes.Buffer
	var outbf bytes.Buffer

	resolution := smf.MetricTicks(1920)
	bpm := 120.00

	wr := writer.NewSMF(&outbf, 1, smfwriter.TimeFormat(resolution))
	wr.WriteHeader()
	wr.Write(meta.FractionalBPM(bpm)) // set the initial bpm
	wr.Write(meter.M4_4())            // set the meter if needed

	rd := midireader.New(&inbf, nil)

	ch := make(chan timedMsg)
	bpmCh := make(chan float64) // allows to change bpm on the fly
	internalStop := make(chan struct{})

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		for {
			select {

			case bpm = <-bpmCh: // change the bpm
				wr.Write(meta.FractionalBPM(bpm))

			case tm := <-ch:
				deltaticks := resolution.FractionalTicks(bpm, time.Duration(tm.deltaMicrosecs)*time.Microsecond)

				wr.SetDelta(deltaticks)
				inbf.Write(tm.data)
				msg, _ := rd.Read()
				wr.Write(msg)
			case <-internalStop:
				wg.Done()
				return
			}
		}
	}()

	in.SetListener(func(data []byte, deltaMicrosecs int64) {
		if len(data) == 0 {
			return
		}
		ch <- timedMsg{data: data, deltaMicrosecs: deltaMicrosecs}
	})

	// wait until we're told to stop
	<-stop
	in.StopListening()
	internalStop <- struct{}{}
	wg.Wait()

	wr.Write(meta.EndOfTrack)
	// get the epoch
	file := fmt.Sprintf("recordings/%d.mid", time.Now().Unix())
	log.Println("Writing to", file)
	ioutil.WriteFile(file, outbf.Bytes(), 0644)

}
