// Package logger provides logging functionality for the GHI application.
// It supports debug logging to date-rotated files in the ~/.ghi/logs directory.
package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	// IsDebug indicates whether debug mode is enabled
	IsDebug      bool
	debugEnabled bool
	debugLogger  *log.Logger
)

// init initializes the logger
func init() {
	// Get the user's home directory
	home, err := os.UserHomeDir()
	if err != nil {
		log.Printf("Warning: Could not get home directory: %v", err)
		return
	}

	// Create the .ghi/logs directory if it doesn't exist
	logDir := filepath.Join(home, ".ghi", "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Printf("Warning: Could not create log directory: %v", err)
		return
	}

	// Create or open the log file with today's date
	logFile := filepath.Join(logDir, fmt.Sprintf("ghi-%s.log", time.Now().Format("2006-01-02")))
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Warning: Could not open log file: %v", err)
		return
	}

	debugLogger = log.New(f, "", log.Ldate|log.Ltime|log.Lmicroseconds)
}

// SetupLogging configures logging based on whether debug mode is enabled.
// If debug is true, logs will be written to date-rotated files in ~/.ghi/logs.
// Otherwise, logs will continue to go to stderr.
func SetupLogging(debug bool) {
	IsDebug = debug
	if !debug {
		return // Keep default stderr logging
	}

	// Create logs directory if it doesn't exist
	logsDir := filepath.Join(getUserHomeDir(), ".ghi", "logs")
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		log.Printf("Warning: Failed to create logs directory: %v", err)
		return
	}

	// Set up the log file with rotation
	logFileName := filepath.Join(logsDir, fmt.Sprintf("ghi-%s.log", time.Now().Format("2006-01-02")))

	// Configure lumberjack for log rotation
	logRotator := &lumberjack.Logger{
		Filename:   logFileName,
		MaxSize:    10, // megabytes
		MaxBackups: 5,
		MaxAge:     30, // days
		Compress:   true,
	}

	// Set the log output to the rotated file
	log.SetOutput(logRotator)
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)

	log.Println("Debug logging enabled")
}

// SetDebug enables or disables debug logging
func SetDebug(enabled bool) {
	debugEnabled = enabled
	if enabled {
		Debug("Debug logging enabled")
	}
}

// Debug logs a message if debug mode is enabled
func Debug(format string, v ...interface{}) {
	if IsDebug {
		log.Printf(format, v...)
	}
	if debugEnabled && debugLogger != nil {
		// Get the source file and line number
		_, file := filepath.Split(os.Args[0])
		debugLogger.Printf("%s:%d: %s", file, 58, fmt.Sprintf(format, v...))
	}
}

// getUserHomeDir returns the user's home directory
func getUserHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Printf("Warning: Failed to determine user home directory: %v", err)
		return "."
	}
	return home
}
