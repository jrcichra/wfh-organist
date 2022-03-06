package timer

import (
	"time"
)

// A very simple timer because I'm concerned that timer.Reset() is not going to like my use case here
type Timer struct {
	requestedSeconds int
	secondsLeft      int
	resetChan        chan struct{}
	doneChan         chan struct{}
}

func (t *Timer) New(seconds int) chan struct{} {
	t.requestedSeconds = seconds
	t.secondsLeft = seconds
	t.resetChan = make(chan struct{}, 10) // don't block on a reset message if we're sleeping
	t.doneChan = make(chan struct{})
	return t.doneChan
}

func (t *Timer) Start() {
	go func() {
		for {
			select {
			// see if a reset is requested
			case <-t.resetChan:
				// reset the timer
				t.secondsLeft = t.requestedSeconds
			default:
			}

			if t.secondsLeft <= 0 {
				t.doneChan <- struct{}{}
				return
			}

			// Sleep and subtract a second
			time.Sleep(1 * time.Second)
			t.secondsLeft -= 1
		}
	}()
}

func (t *Timer) Reset() {
	t.resetChan <- struct{}{}
}
