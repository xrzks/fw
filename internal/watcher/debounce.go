package watcher

import (
	"context"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Debouncer struct {
	cancel  context.CancelFunc
	eventCh chan struct{}
	wg      sync.WaitGroup
	mu      sync.Mutex
	event   fsnotify.Event
}

func NewDebouncer(ctx context.Context, delay time.Duration, callback func(fsnotify.Event)) *Debouncer {
	ctx, cancel := context.WithCancel(ctx)
	eventCh := make(chan struct{}, 1)
	d := &Debouncer{
		cancel:  cancel,
		eventCh: eventCh,
	}

	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case <-eventCh:
				timer := time.NewTimer(delay)
			Drain:
				for {
					select {
					case <-ctx.Done():
						timer.Stop()
						return
					case <-eventCh:
						if !timer.Stop() {
							select {
							case <-timer.C:
							default:
							}
						}
						timer.Reset(delay)
					case <-timer.C:
						break Drain
					}
				}
				d.mu.Lock()
				event := d.event
				d.mu.Unlock()
				callback(event)
			}
		}
	}()

	return d
}

func (d *Debouncer) Trigger(event fsnotify.Event) {
	d.mu.Lock()
	d.event = event
	d.mu.Unlock()

	select {
	case d.eventCh <- struct{}{}:
	default:
	}
}

func (d *Debouncer) Stop() {
	d.cancel()
	d.wg.Wait()
}
