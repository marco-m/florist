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

type Flower struct {
	FilesFS fs.FS
	Port    int `default:"22"`
	// FIXME mmmh I ma not sure I want the paths to be overridable...
	SrcSshdConfigPath        string `default:"sshd/sshd_config.tmpl"`
	DstSshdConfigPath        string `default:"/etc/ssh/sshd_config"`
	SshHostEd25519KeyPub     string `default:"CHANGEME" json:"ssh_host_ed25519_key_pub"`
	SshHostEd25519Key        string `default:"CHANGEME" json:"ssh_host_ed25519_key"`
	SshHostEd25519KeyCertPub string `default:"CHANGEME" json:"ssh_host_ed25519_key_cert_pub"`
	log                      hclog.Logger
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

	root, err := user.Current()
	if err != nil {
		return fmt.Errorf("%s.install: %s", fl, err)
	}

	fl.log.Info("Install sshd configuration file")
	if err := florist.CopyFileTemplateFromFs(fl.FilesFS,
		fl.SrcSshdConfigPath, fl.DstSshdConfigPath,
		0644, root, fl); err != nil {
		return fmt.Errorf("%s.install: %s", fl, err)
	}

	return nil
}

func (fl *Flower) Configure() error {
	fl.log.Info("begin")
	defer fl.log.Info("end")

	root, err := user.Current()
	if err != nil {
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

	fl.log.Info("Adding SSH host key, private")
	if err := florist.CopyFileTemplateFromFs(fl.FilesFS,
		"sshd/ssh_host_ed25519_key.tmpl", "/etc/ssh/ssh_host_ed25519_key",
		0400, root, fl); err != nil {
		return fmt.Errorf("%s.configure: %s", fl, err)
	}
	fl.log.Info("Adding SSH host key, public")
	if err := florist.CopyFileTemplateFromFs(fl.FilesFS,
		"sshd/ssh_host_ed25519_key.pub.tmpl", "/etc/ssh/ssh_host_ed25519_key.pub",
		0400, root, fl); err != nil {
		return fmt.Errorf("%s.configure: %s", fl, err)
	}
	fl.log.Info("Adding SSH host key, certificate")
	if err := florist.CopyFileTemplateFromFs(fl.FilesFS,
		"sshd/ssh_host_ed25519_key-cert.pub.tmpl", "/etc/ssh/ssh_host_ed25519_key-cert.pub",
		0400, root, fl); err != nil {
		return fmt.Errorf("%s.configure: %s", fl, err)
	}

	// Test mode. Only check the validity of the configuration file and sanity of the keys.
	// This gives better diagnostics in case of error.
	fl.log.Info("Checking validity of configuration file")
	cmd := exec.Command("/usr/sbin/sshd", "-t")
	if err := florist.CmdRun(fl.log, cmd); err != nil {
		return fmt.Errorf("%s.configure: check sshd configuration: %s", fl, err)
	}

	fl.log.Info("Reloading sshd service")
	if err := systemd.Reload("ssh"); err != nil {
		return fmt.Errorf("%s.configure: %s", fl, err)
	}

	return nil
}
