package watcher

import (
	"time"

	"github.com/fsnotify/fsnotify"
)

type Debouncer struct {
	timer    *time.Timer
	delay    time.Duration
	callback func(fsnotify.Event)
	event    fsnotify.Event
	hasEvent bool
}

func NewDebouncer(delay time.Duration, callback func(fsnotify.Event)) *Debouncer {
	return &Debouncer{
		// one hour to prevent premature firing
		timer:    time.NewTimer(time.Hour),
		delay:    delay,
		callback: callback,
	}
}

func (d *Debouncer) Trigger(event fsnotify.Event) {
	if !d.timer.Stop() {
		<-d.timer.C
	}
	d.event = event
	d.hasEvent = true
	d.timer.Reset(d.delay)
}

func (d *Debouncer) Stop() {
	if !d.timer.Stop() {
		select {
		case <-d.timer.C:
		default:
		}
	}
}

func (d *Debouncer) getTimerChannel() <-chan time.Time {
	return d.timer.C
}

func (d *Debouncer) onTimerFire() {
	if d.hasEvent {
		d.callback(d.event)
		d.hasEvent = false
	}
}
