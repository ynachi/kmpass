package app

import (
	"golang.org/x/exp/slog"
	"os"
)

// Logger Application wide logger
var Logger = slog.New(slog.NewJSONHandler(os.Stdout))

// init logger
func init() {
	slog.SetDefault(Logger)
}
