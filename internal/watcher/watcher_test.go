package watcher

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/charmbracelet/log"
	"github.com/fsnotify/fsnotify"
	"github.com/xrzks/fw/internal/ignore"
)

func newTestLogger() *log.Logger {
	logger := log.New(os.Stderr)
	logger.SetLevel(log.InfoLevel)
	return logger
}

func newTestIgnorer() *ignore.Matcher {
	return ignore.New(nil)
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
		errCh <- Watch(ctx, dir, commands, nil, newTestIgnorer(), newTestLogger())
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
		errCh <- Watch(ctx, dir, commands, nil, newTestIgnorer(), newTestLogger())
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

	err := Watch(ctx, "/nonexistent/path/that/does/not/exist", nil, nil, newTestIgnorer(), newTestLogger())
	if err == nil {
		t.Error("expected error for invalid path")
	}
}

func TestWatchContextCancellation(t *testing.T) {
	dir := t.TempDir()

	ctx, cancel := context.WithCancel(context.Background())

	errCh := make(chan error, 1)
	go func() {
		errCh <- Watch(ctx, dir, []string{"echo test"}, nil, newTestIgnorer(), newTestLogger())
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

func TestWatchRecursiveDetectsNestedFileChange(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "a", "b")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatalf("failed to create nested dirs: %v", err)
	}

	nestedFile := filepath.Join(sub, "deep.txt")
	if err := os.WriteFile(nestedFile, []byte("initial"), 0o644); err != nil {
		t.Fatalf("failed to create nested file: %v", err)
	}

	marker := filepath.Join(dir, "ran")
	commands := []string{"touch " + marker}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- Watch(ctx, dir, commands, nil, newTestIgnorer(), newTestLogger())
	}()

	time.Sleep(200 * time.Millisecond)

	if err := os.WriteFile(nestedFile, []byte("modified"), 0o644); err != nil {
		t.Fatalf("failed to modify nested file: %v", err)
	}

	err := waitForFile(marker, 3*time.Second)
	if err != nil {
		t.Fatalf("command did not execute after nested file change: %v", err)
	}

	cancel()

	if err := <-errCh; err != nil {
		t.Fatalf("Watch returned error: %v", err)
	}
}

func TestWatchRecursiveDetectsNewDirectory(t *testing.T) {
	dir := t.TempDir()

	marker := filepath.Join(dir, "ran")
	commands := []string{"touch " + marker}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- Watch(ctx, dir, commands, nil, newTestIgnorer(), newTestLogger())
	}()

	time.Sleep(200 * time.Millisecond)

	newDir := filepath.Join(dir, "subdir")
	if err := os.Mkdir(newDir, 0o755); err != nil {
		t.Fatalf("failed to create new directory: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	newFile := filepath.Join(newDir, "file.txt")
	if err := os.WriteFile(newFile, []byte("hello"), 0o644); err != nil {
		t.Fatalf("failed to create file in new directory: %v", err)
	}

	err := waitForFile(marker, 3*time.Second)
	if err != nil {
		t.Fatalf("command did not execute after file change in newly created directory: %v", err)
	}

	cancel()

	if err := <-errCh; err != nil {
		t.Fatalf("Watch returned error: %v", err)
	}
}

func TestAddWatchDir(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "a", "b"), 0o755); err != nil {
		t.Fatalf("failed to create nested dirs: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "a", "file.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		t.Fatalf("failed to create watcher: %v", err)
	}
	defer watcher.Close()

	logger := newTestLogger()

	count, err := addWatchDir(watcher, dir, ignore.New(nil), logger)
	if err != nil {
		t.Fatalf("addWatchDir failed: %v", err)
	}

	if count != 3 {
		t.Errorf("expected 3 directories watched, got %d", count)
	}
}

func TestGetPathTypeInvalid(t *testing.T) {
	_, err := getPathType("/nonexistent/path")
	if err == nil {
		t.Error("expected error for invalid path")
	}
}

func TestMatchExtension(t *testing.T) {
	tests := []struct {
		name       string
		fileName   string
		extensions []string
		want       bool
	}{
		{"no extensions matches all", "foo.go", nil, true},
		{"empty extensions matches all", "foo.go", []string{}, true},
		{"exact match with dot", "foo.go", []string{".go"}, true},
		{"match without dot", "foo.go", []string{"go"}, true},
		{"case insensitive", "foo.Go", []string{".go"}, true},
		{"no match", "foo.go", []string{".rs"}, false},
		{"multiple extensions match", "foo.go", []string{".rs", ".go"}, true},
		{"multiple extensions no match", "foo.go", []string{".rs", ".py"}, false},
		{"no extension file", "Makefile", []string{".go"}, false},
		{"no extension file nil filter", "Makefile", nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchExtension(tt.fileName, tt.extensions); got != tt.want {
				t.Errorf("matchExtension(%q, %v) = %v, want %v", tt.fileName, tt.extensions, got, tt.want)
			}
		})
	}
}

func TestWatchExtensionFilter(t *testing.T) {
	dir := t.TempDir()
	goFile := filepath.Join(dir, "test.go")
	txtFile := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(goFile, []byte("initial"), 0o644); err != nil {
		t.Fatalf("failed to create go file: %v", err)
	}
	if err := os.WriteFile(txtFile, []byte("initial"), 0o644); err != nil {
		t.Fatalf("failed to create txt file: %v", err)
	}

	marker := filepath.Join(dir, "ran")
	commands := []string{"touch " + marker}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- Watch(ctx, dir, commands, []string{".go"}, newTestIgnorer(), newTestLogger())
	}()

	time.Sleep(200 * time.Millisecond)

	if err := os.WriteFile(txtFile, []byte("modified"), 0o644); err != nil {
		t.Fatalf("failed to modify txt file: %v", err)
	}

	time.Sleep(1 * time.Second)
	if _, err := os.Stat(marker); err == nil {
		t.Fatal("command should not have run for .txt file when filtering by .go")
	}

	if err := os.WriteFile(goFile, []byte("modified"), 0o644); err != nil {
		t.Fatalf("failed to modify go file: %v", err)
	}

	if err := waitForFile(marker, 3*time.Second); err != nil {
		t.Fatalf("command did not execute for .go file: %v", err)
	}

	cancel()
	if err := <-errCh; err != nil {
		t.Fatalf("Watch returned error: %v", err)
	}
}

func TestWatchIgnoredFileSkipped(t *testing.T) {
	dir := t.TempDir()
	ignoredDir := filepath.Join(dir, "node_modules", "pkg")
	if err := os.MkdirAll(ignoredDir, 0o755); err != nil {
		t.Fatalf("failed to create ignored dir: %v", err)
	}

	ignoredFile := filepath.Join(ignoredDir, "index.js")
	if err := os.WriteFile(ignoredFile, []byte("initial"), 0o644); err != nil {
		t.Fatalf("failed to create ignored file: %v", err)
	}

	marker := filepath.Join(dir, "ran")
	commands := []string{"touch " + marker}
	ignorer := ignore.New([]string{"node_modules"})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- Watch(ctx, dir, commands, nil, ignorer, newTestLogger())
	}()

	time.Sleep(200 * time.Millisecond)

	if err := os.WriteFile(ignoredFile, []byte("modified"), 0o644); err != nil {
		t.Fatalf("failed to modify ignored file: %v", err)
	}

	time.Sleep(1 * time.Second)
	if _, err := os.Stat(marker); err == nil {
		t.Fatal("command should not have run for ignored file")
	}

	cancel()
	if err := <-errCh; err != nil {
		t.Fatalf("Watch returned error: %v", err)
	}
}

func TestWatchIgnoredNewDirectorySkipped(t *testing.T) {
	dir := t.TempDir()

	marker := filepath.Join(dir, "ran")
	commands := []string{"touch " + marker}
	ignorer := ignore.New([]string{"node_modules"})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- Watch(ctx, dir, commands, nil, ignorer, newTestLogger())
	}()

	time.Sleep(200 * time.Millisecond)

	newDir := filepath.Join(dir, "node_modules", "pkg")
	if err := os.MkdirAll(newDir, 0o755); err != nil {
		t.Fatalf("failed to create new dir: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	newFile := filepath.Join(newDir, "index.js")
	if err := os.WriteFile(newFile, []byte("hello"), 0o644); err != nil {
		t.Fatalf("failed to create file in ignored dir: %v", err)
	}

	time.Sleep(1 * time.Second)
	if _, err := os.Stat(marker); err == nil {
		t.Fatal("command should not have run for file in ignored directory")
	}

	cancel()
	if err := <-errCh; err != nil {
		t.Fatalf("Watch returned error: %v", err)
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
