package app

import (
	"golang.org/x/exp/slog"
	"os"
	"os/exec"
)

// Logger Application wide logger
var Logger = slog.New(slog.NewJSONHandler(os.Stdout))

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
