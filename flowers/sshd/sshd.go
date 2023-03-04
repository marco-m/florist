// Package sshd contains a flower to configure the OpenSSH sshd server.
//
// To generate host keys, see README.
package sshd

import (
	"fmt"
	"io/fs"
	"os/exec"
	"os/user"

	"github.com/creasty/defaults"
	"github.com/hashicorp/go-hclog"

	"github.com/marco-m/florist"
	"github.com/marco-m/florist/pkg/systemd"
)

const (
	ConfigPathSrc = "files/sshd/sshd_config.tmpl"
	ConfigPathDst = "/etc/ssh/sshd_config"

	SshHostEd25519KeySecret = "secrets/sshd/ssh_host_ed25519_key"
	SshHostEd25519KeySrc    = "files/sshd/ssh_host_ed25519_key.tmpl"
	SshHostEd25519KeyDst    = "/etc/ssh/ssh_host_ed25519_key"

	SshHostEd25519KeyPubSecret = "secrets/sshd/ssh_host_ed25519_key.pub"
	SshHostEd25519KeyPubSrc    = "files/sshd/ssh_host_ed25519_key.pub.tmpl"
	SshHostEd25519KeyPubDst    = "/etc/ssh/ssh_host_ed25519_key.pub"

	SshHostEd25519KeyCertPubSecret = "secret/sshd/ssh_host_ed25519_key-cert.pub"
	SshHostEd25519KeyCertPubSrc    = "files/sshd/ssh_host_ed25519_key-cert.pub.tmpl"
	SshHostEd25519KeyCertPubDst    = "/etc/ssh/ssh_host_ed25519_key-cert.pub"
)

var _ florist.Flower = (*Flower)(nil)

type Flower struct {
	fsys fs.FS
	Port int `default:"22"`
	log  hclog.Logger
}

func (fl *Flower) String() string {
	return "florist.flower.sshd"
}

func (fl *Flower) Description() string {
	return "configure an already-existing sshd server"
}

func (fl *Flower) Init(fsys fs.FS) error {
	fl.fsys = fsys
	fl.log = florist.Log.ResetNamed(fl.String())

	if err := defaults.Set(fl); err != nil {
		return fmt.Errorf("%s.init: %s", fl, err)
	}

	return nil
}

func (fl *Flower) Install() error {
	log := fl.log.Named("install")

	root, err := user.Current()
	if err != nil {
		return fmt.Errorf("%s.install: %s", fl, err)
	}

	log.Info("installing sshd configuration file")
	data := map[string]any{"Port": fl.Port}
	if err := florist.CopyFileTemplateFromFs(fl.fsys, ConfigPathSrc, ConfigPathDst,
		0644, root, data); err != nil {
		return fmt.Errorf("%s.install: %s", fl, err)
	}

	return nil
}

func (fl *Flower) Configure() error {
	log := fl.log.Named("configure")

	root, err := user.Current()
	if err != nil {
		return fmt.Errorf("%s.configure: %s", fl, err)
	}

	// FIXME This is dangerous when developing. Is it worthwhile?
	// log.Info("Remove SSH keys already present")
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

	log.Debug("loading secrets")
	data, err := florist.MakeTmplData(fl.fsys,
		SshHostEd25519KeySecret,
		SshHostEd25519KeyPubSecret,
		SshHostEd25519KeyCertPubSecret)
	if err != nil {
		return fmt.Errorf("%s:\n%s", log.Name(), err)
	}

	log.Info("adding SSH host key, private")
	if err := florist.CopyFileTemplateFromFs(fl.fsys,
		SshHostEd25519KeySrc, SshHostEd25519KeyDst, 0400, root, data); err != nil {
		return fmt.Errorf("%s.configure: %s", fl, err)
	}
	log.Info("adding SSH host key, public")
	if err := florist.CopyFileTemplateFromFs(fl.fsys,
		SshHostEd25519KeyPubSrc, SshHostEd25519KeyPubDst, 0400, root, data); err != nil {
		return fmt.Errorf("%s.configure: %s", fl, err)
	}
	log.Info("adding SSH host key, certificate")
	if err := florist.CopyFileTemplateFromFs(fl.fsys,
		SshHostEd25519KeyCertPubSrc, SshHostEd25519KeyCertPubDst,
		0400, root, data); err != nil {
		return fmt.Errorf("%s.configure: %s", fl, err)
	}

	// Flag -t only checks the validity of the configuration file and sanity of the keys.
	// This gives better diagnostics in case of error.
	log.Info("checking validity of configuration file")
	cmd := exec.Command("/usr/sbin/sshd", "-t")
	if err := florist.CmdRun(log, cmd); err != nil {
		return fmt.Errorf("%s.configure: check sshd configuration: %s", fl, err)
	}

	log.Info("reloading sshd service")
	if err := systemd.Reload("ssh"); err != nil {
		return fmt.Errorf("%s.configure: %s", fl, err)
	}

	return nil
}
