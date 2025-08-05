package apt

import (
	"fmt"
	"log/slog"
	"os/exec"

	"github.com/marco-m/florist/pkg/florist"
)

func DpkgInstall(pkgPath string) error {
	log := slog.With("fn", "apt.DpkgInstall")
	log.Info("Installing", "package", pkgPath)

	cmd := exec.Command("dpkg", "--install", pkgPath)
	if err := florist.CmdRun(log, cmd); err != nil {
		return fmt.Errorf("dpkg: install: %s", err)
	}

	return nil
}
