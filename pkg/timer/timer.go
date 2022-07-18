package timer

import (
	"context"
	"time"
)

// A timer that supports resetting without impacting consumers of the timer.
type Timer struct {
	seconds int64
	context context.Context
	cancel  context.CancelFunc
	reset   chan struct{}
}

func NewTimer(duration time.Duration) *Timer {
	t := &Timer{}
	t.seconds = int64(duration.Seconds())
	t.context, t.cancel = context.WithCancel(context.Background())
	t.reset = make(chan struct{}, 1)
	return t
}

func (t *Timer) Done() <-chan struct{} {
	return t.context.Done()
}

func (t *Timer) Start() {
	remaining := t.seconds
	go func() {
		for {
			select {
			case <-t.reset:
				remaining = t.seconds
			case <-t.context.Done():
				return
			default:
				time.Sleep(1 * time.Second)
				remaining--
				// log.Println("Remaining:", remaining)
				if remaining <= 0 {
					// log.Println("Called cancel for the timer because remaining is 0")
					t.cancel()
				}
			}
		}
	}()
}

func (t *Timer) Reset() {
	// log.Println("context err for timer", t.context.Err())
	if t.context.Err() == nil {
		// reset the still running timer
		select {
		case t.reset <- struct{}{}:
		default:
		}
	} else {
		// reset the context if the context completed
		t.context, t.cancel = context.WithCancel(context.Background())
	}
}

func (t *Timer) Stop() {
	// log.Println("Called cancel for the timer from stop")
	t.cancel()
}
