package systemd

import (
	"fmt"
	"os/exec"

	"github.com/marco-m/florist/pkg/florist"
)

func Enable(unit string) error {
	log := florist.Log().With("pkg", "systemd").With("unit", unit)

	cmd := exec.Command("systemctl", "enable", unit)
	if err := florist.CmdRun(log, cmd); err != nil {
		return fmt.Errorf("florist.systemd: enable: %s", err)
	}
	return nil
}

func Start(unit string) error {
	log := florist.Log().With("pkg", "systemd")

	cmd := exec.Command("systemctl", "start", unit)
	if err := florist.CmdRun(log, cmd); err != nil {
		return fmt.Errorf("florist.systemd: start: %s", err)
	}
	return nil
}

func Restart(unit string) error {
	log := florist.Log().With("pkg", "systemd")

	cmd := exec.Command("systemctl", "restart", unit)
	if err := florist.CmdRun(log, cmd); err != nil {
		return fmt.Errorf("florist.systemd: restart: %s", err)
	}
	return nil
}

func Reload(unit string) error {
	log := florist.Log().With("pkg", "systemd")

	cmd := exec.Command("systemctl", "reload", unit)
	if err := florist.CmdRun(log, cmd); err != nil {
		return fmt.Errorf("florist.systemd: reload: %s", err)
	}
	return nil
}

// Status executes "systemctl status unit".
// WARNING: in case the unit is stopped, Status will return an error.
// There is a set of status code, that I might translate to Go errors.
func Status(unit string) error {
	log := florist.Log().With("pkg", "systemd")

	var cmd *exec.Cmd
	if unit != "" {
		cmd = exec.Command("systemctl", "status", "--lines=0", unit)
	} else {
		cmd = exec.Command("systemctl", "status")
	}

	if err := florist.CmdRun(log, cmd); err != nil {
		return fmt.Errorf("florist.systemd: status: %s", err)
	}
	return nil
}
