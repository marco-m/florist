package systemd

import (
	"fmt"
	"os/exec"

	"github.com/marco-m/florist"
)

func Enable(unit string) error {
	log := florist.Log.Named("systemd").With("unit", unit)

	cmd := exec.Command("systemctl", "enable", unit)
	if err := florist.CmdRun(log, cmd); err != nil {
		return fmt.Errorf("florist.systemd: enable: %s", err)
	}
	return nil
}

func Start(unit string) error {
	log := florist.Log.Named("systemd")

	cmd := exec.Command("systemctl", "start", unit)
	if err := florist.CmdRun(log, cmd); err != nil {
		return fmt.Errorf("florist.systemd: start: %s", err)
	}
	return nil
}
