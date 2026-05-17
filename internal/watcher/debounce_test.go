package watcher

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
)

func TestDebouncerFiresAfterDelay(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	called := make(chan struct{}, 1)
	d := NewDebouncer(ctx, 50*time.Millisecond, func(event fsnotify.Event) {
		select {
		case called <- struct{}{}:
		default:
		}
	})

	d.Trigger(fsnotify.Event{Name: "test.txt", Op: fsnotify.Write})

	select {
	case <-called:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("expected callback to be called, timed out")
	}

	d.Stop()
}

func TestDebouncerCoalescesMultipleTriggers(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	called := make(chan struct{}, 1)
	d := NewDebouncer(ctx, 100*time.Millisecond, func(event fsnotify.Event) {
		select {
		case called <- struct{}{}:
		default:
		}
	})

	for i := 0; i < 10; i++ {
		d.Trigger(fsnotify.Event{Name: "test.txt", Op: fsnotify.Write})
		time.Sleep(10 * time.Millisecond)
	}

	select {
	case <-called:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("expected debounced callback to fire, timed out")
	}

	time.Sleep(150 * time.Millisecond)

	select {
	case <-called:
		t.Error("expected only one callback, got a second")
	default:
	}

	d.Stop()
}

func TestDebouncerPassesLastEvent(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	eventCh := make(chan fsnotify.Event, 1)
	d := NewDebouncer(ctx, 50*time.Millisecond, func(event fsnotify.Event) {
		select {
		case eventCh <- event:
		default:
		}
	})

	d.Trigger(fsnotify.Event{Name: "first.txt", Op: fsnotify.Create})
	time.Sleep(10 * time.Millisecond)
	d.Trigger(fsnotify.Event{Name: "last.txt", Op: fsnotify.Write})

	var lastEvent fsnotify.Event
	select {
	case lastEvent = <-eventCh:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("expected callback to fire, timed out")
	}

	if lastEvent.Name != "last.txt" {
		t.Errorf("expected event name 'last.txt', got %q", lastEvent.Name)
	}
	if lastEvent.Op != fsnotify.Write {
		t.Errorf("expected event op Write, got %v", lastEvent.Op)
	}

	d.Stop()
}

func TestDebouncerFiresMultipleTimes(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	called := make(chan struct{}, 1)
	d := NewDebouncer(ctx, 50*time.Millisecond, func(event fsnotify.Event) {
		select {
		case called <- struct{}{}:
		default:
		}
	})

	d.Trigger(fsnotify.Event{Name: "a.txt", Op: fsnotify.Write})
	select {
	case <-called:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("expected first callback, timed out")
	}

	d.Trigger(fsnotify.Event{Name: "b.txt", Op: fsnotify.Write})
	select {
	case <-called:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("expected second callback, timed out")
	}

	d.Stop()
}

func TestDebouncerStop(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var called atomic.Int32
	calledCh := make(chan struct{}, 1)
	d := NewDebouncer(ctx, 50*time.Millisecond, func(event fsnotify.Event) {
		called.Add(1)
		select {
		case calledCh <- struct{}{}:
		default:
		}
	})

	d.Trigger(fsnotify.Event{Name: "test.txt", Op: fsnotify.Write})
	select {
	case <-calledCh:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("expected callback before stop, timed out")
	}

	d.Stop()

	d.Trigger(fsnotify.Event{Name: "after.txt", Op: fsnotify.Write})
	time.Sleep(150 * time.Millisecond)

	if got := called.Load(); got != 1 {
		t.Errorf("expected no more callbacks after stop, got %d", got)
	}
}

func TestDebouncerContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	var called atomic.Int32
	d := NewDebouncer(ctx, 200*time.Millisecond, func(event fsnotify.Event) {
		called.Add(1)
	})

	d.Trigger(fsnotify.Event{Name: "test.txt", Op: fsnotify.Write})
	time.Sleep(20 * time.Millisecond)

	cancel()

	d.Stop()

	if got := called.Load(); got != 0 {
		t.Errorf("expected no callbacks after context cancellation, got %d", got)
	}
}
