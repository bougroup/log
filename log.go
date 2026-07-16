package log

import (
	"fmt"
	golog "log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

// logMu guards the logger/verbosity configuration below so SetLogLevel and
// SetLogger can be called safely while other goroutines are logging.
var logMu sync.RWMutex
var logger *golog.Logger
var verbosity Level
var verbositySet bool

// Level is used to set verbosity of your app
type Level int

const (
	//LevelNone - log nothing
	LevelNone Level = iota
	//LevelError - logs only errors
	LevelError
	//LevelWarning  - logs warning and errors
	LevelWarning
	//LevelTest - logs only debug messages
	LevelTest
	//LevelInfo - logs info, warining and errors
	LevelInfo
)

// getVerbosity defaults verbosity to LogLevelWarning if a verbosity is not set
func getVerbosity() Level {
	logMu.RLock()
	defer logMu.RUnlock()
	if !verbositySet {
		return LevelWarning
	}
	return verbosity
}

// SetLogLevel sets log level
func SetLogLevel(l Level) {
	logMu.Lock()
	defer logMu.Unlock()
	verbosity = l
	verbositySet = true
}

// SetLogger sets Logger
func SetLogger(l *golog.Logger) {
	logMu.Lock()
	defer logMu.Unlock()
	logger = l
}

// printLine writes an already-formatted line to the configured logger, or to
// stderr when none is set.
func printLine(line string) {
	logMu.RLock()
	l := logger
	logMu.RUnlock()
	if l != nil {
		l.Println(line)
	} else {
		println(line)
	}
}

// caller returns the "file:line" of the logging call site, skip frames up.
func caller(skip int) string {
	_, fn, line, ok := runtime.Caller(skip)
	if !ok {
		return "?"
	}
	return fmt.Sprintf("%s:%d", filepath.Base(fn), line)
}

// Info prints out informational messages
func Info(message string) {
	if getVerbosity() >= LevelInfo {
		emit([]Field{{"time", now()}, {"level", "info"}, {"msg", message}})
	}
}

// Error prints out error messages
func Error(message string) {
	if getVerbosity() >= LevelError {
		emit([]Field{{"time", now()}, {"level", "error"}, {"msg", message}, {"caller", caller(2)}})
	}
}

// Warn prints out warning messages
func Warn(message string) {
	if getVerbosity() >= LevelWarning {
		emit([]Field{{"time", now()}, {"level", "warning"}, {"msg", message}, {"caller", caller(2)}})
	}
}

// Test prints out debugging or test messages
func Test(message string) {
	if getVerbosity() == LevelTest {
		emit([]Field{{"time", now()}, {"level", "debug"}, {"msg", message}, {"caller", caller(2)}})
	}
}

func Fatal(message string) {
	Error(message)
	os.Exit(1)
}
