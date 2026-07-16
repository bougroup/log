package log

import (
	"fmt"
	golog "log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
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

func printLog(message string) {
	logMu.RLock()
	l := logger
	logMu.RUnlock()
	if l != nil {
		l.Println(message)
	} else {
		println("[" + time.Now().Format(time.RFC3339Nano) + "] " + message)
	}
}

// Info prints out informational messages
func Info(message string) {
	if getVerbosity() >= LevelInfo {
		printLog(message)
	}
}

// Error prints out error messages
func Error(message string) {
	if getVerbosity() >= LevelError {
		_, fn, line, _ := runtime.Caller(1)
		printLog(fmt.Sprintf("ERROR: %s (%s:%d)", message, filepath.Base(fn), line))
	}
}

// Warn prints out warning messages
func Warn(message string) {
	if getVerbosity() >= LevelWarning {
		_, fn, line, _ := runtime.Caller(1)
		printLog(fmt.Sprintf("WARNING: %s (%s:%d)", message, filepath.Base(fn), line))
	}
}

// Test prints out debugging or test messages
func Test(message string) {
	if getVerbosity() == LevelTest {
		_, fn, line, _ := runtime.Caller(1)
		printLog(fmt.Sprintf("DEBUG: %s (%s:%d)", message, filepath.Base(fn), line))
	}
}

func Fatal(message string) {
	Error(message)
	os.Exit(1)
}
