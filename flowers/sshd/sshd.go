// Package sshd contains a flower to configure the OpenSSH sshd server.
package sshd

import (
	"fmt"
	"io/fs"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/marco-m/florist"

	"github.com/hashicorp/go-hclog"
)

type Options struct {
	FilesFS fs.FS
}

type Flower struct {
	Options
	log hclog.Logger
}

func New(opts Options) (*Flower, error) {
	fl := Flower{Options: opts}
	name := fmt.Sprintf("florist.flower.%s", fl)
	fl.log = florist.Log.ResetNamed(name)

	if fl.FilesFS == nil {
		return nil, fmt.Errorf("%s.new: missing FilesFS", name)
	}

	return &fl, nil
}

func (fl Flower) String() string {
	return "sshd"
}

func (fl Flower) Description() string {
	return "configure an already-existing sshd server"
}

func (fl *Flower) Install() error {
	fl.log.Info("begin")
	defer fl.log.Info("end")

	root, err := user.Current()
	if err != nil {
		return err
	}

	fl.log.Info("Install sshd configuration file")
	if err := florist.CopyFromFs(fl.FilesFS, "sshd_config",
		"/etc/ssh/sshd_config", 0644, root); err != nil {
		return err
	}

	fl.log.Info("Remove SSH keys already present")
	entries, err := os.ReadDir("/etc/ssh/")
	if err != nil {
		return err
	}
	for _, file := range entries {
		if file.Type().IsRegular() && strings.HasPrefix(file.Name(), "ssh_host_") {
			if err := os.Remove(filepath.Join("/etc/ssh/", file.Name())); err != nil {
				return err
			}
		}
	}

	fl.log.Info("Add SSH host key, private")
	if err := florist.CopyFromFs(fl.FilesFS, "secrets/ssh_host_ed25519_key",
		"/etc/ssh/ssh_host_ed25519_key", 0400, root); err != nil {
		return err
	}
	fl.log.Info("Add SSH host key, public")
	if err := florist.CopyFromFs(fl.FilesFS, "secrets/ssh_host_ed25519_key.pub",
		"/etc/ssh/ssh_host_ed25519_key.pub", 0400, root); err != nil {
		return err
	}
	fl.log.Info("Add SSH host key, certificate")
	if err := florist.CopyFromFs(fl.FilesFS, "secrets/ssh_host_ed25519_key-cert.pub", "/etc/ssh/ssh_host_ed25519_key-cert.pub", 0400, root); err != nil {
		return err
	}

	return nil
}
