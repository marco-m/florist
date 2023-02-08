package apt

import (
	"fmt"
	"os/exec"

	"github.com/marco-m/florist"
)

func DpkgInstall(pkgPath string) error {
	log := florist.Log.Named("apt.DpkgInstall")
	log.Info("Installing", "package", pkgPath)

	cmd := exec.Command("dpkg", "--install", pkgPath)
	if err := florist.CmdRun(log, cmd); err != nil {
		return fmt.Errorf("dpkg: install: %s", err)
	}

	return nil
}