package florist

import "github.com/hashicorp/go-hclog"

// Module-global default logger.
var Log = hclog.NewNullLogger()

// SetLogger configures the logger for the whole florist module. By default no
// logging is emitted.
func SetLogger(log hclog.Logger) {
	Log = log.ResetNamed("florist")
}
