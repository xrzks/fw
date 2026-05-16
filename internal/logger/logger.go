package logger

import (
	"log"
	"os"
)

type Logger struct {
	debug  bool
	logger *log.Logger
}

func New(debug bool) *Logger {
	return &Logger{
		debug:  debug,
		logger: log.New(os.Stdout, "", 0),
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

func (l *Logger) Fatal(format string, v ...any) {
	l.logger.Printf("[FATAL] "+format, v...)
	os.Exit(1)
}

