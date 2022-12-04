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
	FilesFS       fs.FS
	SrcSshdConfig string `default:"sshd/sshd_config.tmpl"`
	DstSshdConfig string `default:"/etc/ssh/sshd_config"`
	Port          int    `default:"22"`
	log           hclog.Logger
}

func (fl *Flower) String() string {
	return "florist.flower.sshd"
}

func (fl *Flower) Description() string {
	return "configure an already-existing sshd server"
}

func (fl *Flower) Init() error {
	fl.log = florist.Log.ResetNamed(fl.String())

	if fl.FilesFS == nil {
		return fmt.Errorf("%s.init: missing FilesFS", fl)
	}

	if err := defaults.Set(fl); err != nil {
		return fmt.Errorf("%s.init: %s", fl, err)
	}

	return nil
}

func (fl *Flower) Install() error {
	fl.log.Info("begin")
	defer fl.log.Info("end")
	return nil
}

func (fl *Flower) Configure() error {
	fl.log.Info("begin")
	defer fl.log.Info("end")

	root, err := user.Current()
	if err != nil {
		return fmt.Errorf("%s.configure: %s", fl, err)
	}

	fl.log.Info("Install sshd configuration file")
	if err := florist.CopyFileTemplateFromFs(fl.FilesFS, fl.SrcSshdConfig,
		fl.DstSshdConfig, 0644, root, fl); err != nil {
		return fmt.Errorf("%s.configure: %s", fl, err)
	}

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

	// FIXME NEED TO THINK A BIT MORE HERE...

	fl.log.Info("Add SSH host key, private")
	if err := florist.CopyFile("secrets/ssh_host_ed25519_key",
		"/etc/ssh/ssh_host_ed25519_key", 0400, root); err != nil {
		return fmt.Errorf("%s.configure: %s", fl, err)
	}
	fl.log.Info("Add SSH host key, public")
	if err := florist.CopyFile("secrets/ssh_host_ed25519_key.pub",
		"/etc/ssh/ssh_host_ed25519_key.pub", 0400, root); err != nil {
		return fmt.Errorf("%s.configure: %s", fl, err)
	}
	fl.log.Info("Add SSH host key, certificate")
	if err := florist.CopyFile("secrets/ssh_host_ed25519_key-cert.pub", "/etc/ssh/ssh_host_ed25519_key-cert.pub", 0400, root); err != nil {
		return fmt.Errorf("%s.configure: %s", fl, err)
	}

	// FIXME NEED TO RELOAD SSHD HERE...

	return nil
}
