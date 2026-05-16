package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

type Logger struct {
	debug  bool
	logger *log.Logger
	file   *os.File
}

func New(debug bool) *Logger {
	l := &Logger{debug: debug}

	var writer io.Writer = os.Stdout
	if debug {
		dir, err := os.UserCacheDir()
		if err == nil {
			dir = filepath.Join(dir, "fw")
			if err := os.MkdirAll(dir, 0o755); err != nil {
				fmt.Fprintf(os.Stderr, "warning: could not create cache directory: %v\n", err)
			} else {
				f, err := os.OpenFile(filepath.Join(dir, "fw-debug.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
				if err == nil {
					writer = io.MultiWriter(os.Stdout, f)
					l.file = f
				} else {
					fmt.Fprintf(os.Stderr, "warning: could not create debug log file: %v\n", err)
				}
			}
		} else {
			fmt.Fprintf(os.Stderr, "warning: could not determine cache directory: %v\n", err)
		}
	}
	l.logger = log.New(writer, "", 0)
	return l
}

func (l *Logger) Close() {
	if l.file != nil {
		if err := l.file.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to close debug log file: %v\n", err)
		}
	}
}

func (l *Logger) Debug(format string, v ...any) {
	if l.debug {
		l.logger.Printf("[DEBUG] "+format, v...)
	}
}

func (l *Logger) Info(format string, v ...any) {
	l.logger.Printf("[INFO] "+format, v...)
}

func (l *Logger) Error(format string, v ...any) {
	l.logger.Printf("[ERROR] "+format, v...)
}
