package log

import (
	"fmt"
	golog "log"
	"path/filepath"
	"runtime"
	"time"
)

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
	if !verbositySet {
		return LevelWarning
	}
	return verbosity
}

// SetLogLevel sets log level
func SetLogLevel(l Level) {
	verbosity = l
	verbositySet = true
}

// SetLogger sets Logger
func SetLogger(l *golog.Logger) {
	logger = l
}

func printLog(message string) {
	if logger != nil {
		logger.Println(message)
	} else {
		println("[" + time.Now().Format("2006-01-02-15:04:05.000000") + "] " + message)
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
