package watcher

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/charmbracelet/log"
)

func newTestLogger() *log.Logger {
	logger := log.New(os.Stderr)
	logger.SetLevel(log.InfoLevel)
	return logger
}

func TestWatchDetectsFileChange(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(file, []byte("initial"), 0o644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	marker := filepath.Join(dir, "ran")
	commands := []string{"touch " + marker}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- Watch(ctx, dir, commands, newTestLogger())
	}()

	time.Sleep(200 * time.Millisecond)

	if err := os.WriteFile(file, []byte("modified"), 0o644); err != nil {
		t.Fatalf("failed to modify test file: %v", err)
	}

	err := waitForFile(marker, 3*time.Second)
	if err != nil {
		t.Fatalf("command did not execute after file change: %v", err)
	}

	cancel()

	if err := <-errCh; err != nil {
		t.Fatalf("Watch returned error: %v", err)
	}
}

func TestWatchDetectsNewFile(t *testing.T) {
	dir := t.TempDir()

	marker := filepath.Join(dir, "ran")
	commands := []string{"touch " + marker}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- Watch(ctx, dir, commands, newTestLogger())
	}()

	time.Sleep(200 * time.Millisecond)

	newFile := filepath.Join(dir, "newfile.txt")
	if err := os.WriteFile(newFile, []byte("new"), 0o644); err != nil {
		t.Fatalf("failed to create new file: %v", err)
	}

	err := waitForFile(marker, 3*time.Second)
	if err != nil {
		t.Fatalf("command did not execute after new file created: %v", err)
	}

	cancel()

	if err := <-errCh; err != nil {
		t.Fatalf("Watch returned error: %v", err)
	}
}

func TestWatchInvalidPath(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := Watch(ctx, "/nonexistent/path/that/does/not/exist", nil, newTestLogger())
	if err == nil {
		t.Error("expected error for invalid path")
	}
}

func TestWatchContextCancellation(t *testing.T) {
	dir := t.TempDir()

	ctx, cancel := context.WithCancel(context.Background())

	errCh := make(chan error, 1)
	go func() {
		errCh <- Watch(ctx, dir, []string{"echo test"}, newTestLogger())
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Watch returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Watch did not stop after context cancellation")
	}
}

func TestGetPathTypeFile(t *testing.T) {
	file, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(file.Name())

	pt, err := getPathType(file.Name())
	if err != nil {
		t.Fatalf("getPathType failed: %v", err)
	}
	if pt != "file" {
		t.Errorf("expected file, got %q", pt)
	}
}

func TestGetPathTypeDir(t *testing.T) {
	dir := t.TempDir()

	pt, err := getPathType(dir)
	if err != nil {
		t.Fatalf("getPathType failed: %v", err)
	}
	if pt != "directory" {
		t.Errorf("expected directory, got %q", pt)
	}
}

func TestGetPathTypeInvalid(t *testing.T) {
	_, err := getPathType("/nonexistent/path")
	if err == nil {
		t.Error("expected error for invalid path")
	}
}

func waitForFile(path string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); err == nil {
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return os.ErrNotExist
}
