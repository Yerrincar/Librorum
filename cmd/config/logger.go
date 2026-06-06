package config

import (
	"encoding/json"
	"io"
	"os"
	"runtime/debug"
	"sync"
	"time"

	"github.com/fatih/color"
)

type Level int8

const (
	LevelDebug Level = iota
	LevelInfo
	LevelError
	LevelFatal
	LevelOff
)

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return color.BlueString("DEBUG")
	case LevelInfo:
		return color.GreenString("INFO")
	case LevelError:
		return color.RedString("ERROR")
	case LevelFatal:
		return color.BlackString("FATAL")
	default:
		return ""
	}
}

type Logger struct {
	out      io.Writer
	minLevel Level
	mu       sync.Mutex
}

func New(out io.Writer, minLevel Level) *Logger {
	return &Logger{
		out:      out,
		minLevel: minLevel,
	}
}

func (l *Logger) Info(message string, properties map[string]string) {
	l.Print(LevelInfo, message, properties)
}
func (l *Logger) Debug(message string, properties map[string]string) {
	l.Print(LevelDebug, message, properties)
}
func (l *Logger) Error(message string, properties map[string]string) {
	l.Print(LevelError, message, properties)
}
func (l *Logger) Fatal(message string, properties map[string]string) {
	l.Print(LevelFatal, message, properties)
	os.Exit(1)
}

func (l *Logger) Print(level Level, message string, properties map[string]string) (int, error) {
	if level < l.minLevel {
		return 0, nil
	}

	aux := struct {
		Level      string            `json:"level"`
		Time       string            `json:"time"`
		Message    string            `json:"message"`
		Properties map[string]string `json:"properties,omitempty"`
		Trace      string            `json:"trace,omitempty"`
	}{
		Level:      level.String(),
		Message:    message,
		Properties: properties,
		Time:       time.Now().UTC().Format(time.RFC3339),
	}

	if level >= LevelError {
		aux.Trace = string(debug.Stack())
	}

	line, err := json.Marshal(aux)
	if err != nil {
		line = []byte(LevelError.String() + ": unable to marshal log message: " + err.Error())
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	return l.out.Write(append(line, '\n'))
}

func (l *Logger) Write(message []byte) (n int, err error) {
	return l.Print(LevelError, string(message), nil)
}
