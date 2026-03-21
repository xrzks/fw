package watcher

import (
	"time"
)

type Debouncer struct {
	timer     *time.Timer
	delay     time.Duration
	callback  func()
}

func NewDebouncer(delay time.Duration, callback func()) *Debouncer {
	return &Debouncer{
		timer:    time.NewTimer(time.Hour),
		delay:    delay,
		callback: callback,
	}
}

func (d *Debouncer) Stop() {
	if !d.timer.Stop() {
		select {
		case <-d.timer.C:
		default:
		}
	}
}

func (d *Debouncer) Trigger() {
	d.Stop()
	d.timer.Reset(d.delay)
}

func (d *Debouncer) Start() <-chan time.Time {
	return d.timer.C
}
