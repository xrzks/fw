package watcher

import (
	"context"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Debouncer struct {
	ctx      context.Context
	timer    *time.Timer
	delay    time.Duration
	callback func(fsnotify.Event)
	mu       sync.Mutex
	event    fsnotify.Event
	hasEvent bool
	doneCh   <-chan time.Time
}

func NewDebouncer(ctx context.Context, delay time.Duration, callback func(fsnotify.Event)) *Debouncer {
	ch := make(chan time.Time)
	close(ch)
	return &Debouncer{
		ctx:      ctx,
		timer:    time.NewTimer(time.Hour),
		delay:    delay,
		callback: callback,
		doneCh:   ch,
	}
}

func (d *Debouncer) Trigger(event fsnotify.Event) {
	select {
	case <-d.ctx.Done():
		return
	default:
	}

	if !d.timer.Stop() {
		select {
		case <-d.timer.C:
		default:
		}
	}
	d.mu.Lock()
	d.event = event
	d.hasEvent = true
	d.mu.Unlock()
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

func (d *Debouncer) TimerChannel() <-chan time.Time {
	select {
	case <-d.ctx.Done():
		return d.doneCh
	default:
		return d.timer.C
	}
}

func (d *Debouncer) OnTimerFire() {
	select {
	case <-d.ctx.Done():
		// Context cancelled, don't execute callback
		return
	default:
	}

	d.mu.Lock()
	event := d.event
	hasEvent := d.hasEvent
	d.hasEvent = false
	d.mu.Unlock()

	if hasEvent {
		d.callback(event)
	}
}
