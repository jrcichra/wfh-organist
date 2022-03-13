package timer

import (
	"time"
)

// A very simple timer because I'm concerned that timer.Reset() is not going to like my use case here
type Timer struct {
	requestedSeconds int64
	secondsLeft      int64
	resetChan        chan struct{}
	doneChan         chan struct{}
	internalDoneChan chan struct{}
}

func (t *Timer) New(seconds int64) chan struct{} {
	t.requestedSeconds = seconds
	t.secondsLeft = seconds
	t.resetChan = make(chan struct{})
	t.doneChan = make(chan struct{})
	return t.doneChan
}

func (t *Timer) Start() {
	go func() {
		for {

			if t.secondsLeft <= 0 {
				t.doneChan <- struct{}{}
				t.internalDoneChan <- struct{}{}
				return
			}
			// Sleep and subtract a second
			time.Sleep(1 * time.Second)
			t.secondsLeft -= 1
		}
	}()

	go func() {
		for {
			// see if a reset is requested
			select {
			case <-t.resetChan:
				// reset the timer
				t.secondsLeft = t.requestedSeconds
			case <-t.internalDoneChan:
				return
			}
		}
	}()

}

func (t *Timer) Reset() {
	t.resetChan <- struct{}{}
}
