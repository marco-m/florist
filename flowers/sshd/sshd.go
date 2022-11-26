// Package sshd contains a flower to configure the OpenSSH sshd server.
package sshd

import (
	"fmt"
	"io/fs"
	"os/user"

	"github.com/marco-m/florist"

	"github.com/creasty/defaults"
	"github.com/hashicorp/go-hclog"
)

type Flower struct {
	FilesFS fs.FS
	Port    int `default:"22"`
	log     hclog.Logger
}

func (fl *Flower) String() string {
	return "sshd"
}

func (fl *Flower) Description() string {
	return "configure an already-existing sshd server"
}

func (fl *Flower) Init() error {
	name := fmt.Sprintf("florist.flower.%s", fl)
	fl.log = florist.Log.ResetNamed(name)

	if fl.FilesFS == nil {
		return fmt.Errorf("%s.init: missing FilesFS", name)
	}

	if err := defaults.Set(fl); err != nil {
		return fmt.Errorf("%s.init: %s", name, err)
	}

	return nil
}

func (fl *Flower) Install() error {
	fl.log.Info("begin")
	defer fl.log.Info("end")

	// FIXME This is dangerous when developing. Is it worthwhile?
	// fl.log.Info("Remove SSH keys already present")
	// entries, err := os.ReadDir("/etc/ssh/")
	// if err != nil {
	// 	return err
	// }
	// for _, file := range entries {
	// 	if file.Type().IsRegular() && strings.HasPrefix(file.Name(), "ssh_host_") {
	// 		if err := os.Remove(filepath.Join("/etc/ssh/", file.Name())); err != nil {
	// 			return err
	// 		}
	// 	}
	// }

	root, err := user.Current()
	if err != nil {
		return err
	}

	fl.log.Info("Install sshd configuration file")
	if err := florist.CopyFileFromFs(fl.FilesFS, "sshd/sshd_config",
		"/etc/ssh/sshd_config", 0644, root); err != nil {
		return err
	}

	// FIXME this should be installed at deployment time
	// fl.log.Info("Add SSH host key, private")
	// if err := florist.CopyFileFromFs(fl.FilesFS, "secrets/ssh_host_ed25519_key",
	// 	"/etc/ssh/ssh_host_ed25519_key", 0400, root); err != nil {
	// 	return err
	// }
	// fl.log.Info("Add SSH host key, public")
	// if err := florist.CopyFileFromFs(fl.FilesFS, "secrets/ssh_host_ed25519_key.pub",
	// 	"/etc/ssh/ssh_host_ed25519_key.pub", 0400, root); err != nil {
	// 	return err
	// }
	// fl.log.Info("Add SSH host key, certificate")
	// if err := florist.CopyFileFromFs(fl.FilesFS, "secrets/ssh_host_ed25519_key-cert.pub", "/etc/ssh/ssh_host_ed25519_key-cert.pub", 0400, root); err != nil {
	// 	return err
	// }

	return nil
}
