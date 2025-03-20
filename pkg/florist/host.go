package florist

import (
	"log/slog"
	"os/exec"
)

// SetHostname attempts to persistently set the hostname to 'name'.
// Many things can go wrong, among them _when_ this function is called.
//
// Testing on AWS Ubuntu instance:
// - call "hostnamectl set-hostname foo"
// - bad:  does not update /etc/hosts.
// - good: overwrites /etc/hostname.
// - good: "host foo" will return the correct IP address.
// - good: survives a reboot.
func SetHostname(log *slog.Logger, name string) error {
	log.Info("SetHostname", "name", name)
	return CmdRun(log, exec.Command("hostnamectl", "set-hostname", name))
}
