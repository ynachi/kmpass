package app

import (
	"golang.org/x/exp/slog"
	"os"
	"os/exec"
)

type LogLevel int

const (
	Info LogLevel = iota
	Warn
	Error
	Debug
)

// Logger Application wide logger
var programLevel = new(slog.LevelVar) // Info by default
var h = slog.HandlerOptions{Level: programLevel}.NewJSONHandler(os.Stdout)
var Logger = slog.New(h)

// init logger
func init() {
	slog.SetDefault(Logger)
	_, err := exec.LookPath("multipass")
	if err != nil {
		Logger.Error("error loading multipass command", err)
		// fail if multipass command is not found as everything is based on
		// it to be functional.
		os.Exit(1)
	}
}

// SetLogLevel sets the application log level dynamically. Defaut level in info.
func SetLogLevel(level LogLevel) {
	switch level {
	case Info:
		programLevel.Set(slog.LevelInfo)
	case Warn:
		programLevel.Set(slog.LevelWarn)
	case Error:
		programLevel.Set(slog.LevelError)
	case Debug:
		programLevel.Set(slog.LevelDebug)
	}
}

