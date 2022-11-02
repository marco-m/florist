package cook

import (
	"os"
	"os/signal"
)

// ConsumeSignals ignores SIGINT, allowing subprocesses time to properly clean up.
// It is quite important to call this function before doing any work.
func ConsumeSignals() {
	go func() {
		signalCh := make(chan os.Signal, 1)
		signal.Notify(signalCh, os.Interrupt)
		for {
			sig := <-signalCh
			Out("ignoring received signal", sig)
		}
	}()
}
