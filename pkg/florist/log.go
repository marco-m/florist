package florist

import (
	"os"

	"github.com/hashicorp/go-hclog"
)

// Module-global default logger.
var Log = hclog.NewNullLogger()

// name should be <project>-florist
func NewLogger(name string) hclog.Logger {
	return hclog.New(&hclog.LoggerOptions{
		Name:       name,
		Level:      hclog.LevelFromString("DEBUG"),
		TimeFormat: "15:04:05Z07:00",
		// color output is unreadable on displays with a light theme :-(
		Color: hclog.ColorOff,
		// the default output is os.Stderr, which is rendered by Packer in red,
		// with the effect of making it seem a long stream of errors :-/
		Output: os.Stdout,
	})
}

// SetLogger configures the logger for the whole florist module.
// By default no logging is emitted.
func SetLogger(log hclog.Logger) {
	Log = log.ResetNamed("florist")
}
