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

	"github.com/marco-m/florist/pkg/florist"
	"github.com/marco-m/florist/pkg/systemd"
)

type SrcDst struct {
	Src string
	Dst string
}

var SshdConfig = SrcDst{
	Src: "sshd_config.tpl",
	Dst: "/etc/ssh/sshd_config",
}

var SshHostEd25519Key = SrcDst{
	Src: "ssh_host_ed25519_key.tpl",
	Dst: "/etc/ssh/ssh_host_ed25519_key",
}

var SshHostEd25519KeyPub = SrcDst{
	Src: "ssh_host_ed25519_key.pub.tpl",
	Dst: "/etc/ssh/ssh_host_ed25519_key.pub",
}

var SshHostEd25519KeyCertPub = SrcDst{
	Src: "ssh_host_ed25519_key-cert.pub.tpl",
	Dst: "/etc/ssh/ssh_host_ed25519_key-cert.pub",
}

type secrets struct {
	SshHostEd25519Key        string
	SshHostEd25519KeyPub     string
	SshHostEd25519KeyCertPub string
}

var _ florist.Flower = (*Flower)(nil)

type Flower struct {
	Port int `default:"22"`

	log hclog.Logger
}

func (fl *Flower) String() string {
	return "sshd"
}

func (fl *Flower) Description() string {
	return "configure an already-existing sshd server"
}

func (fl *Flower) Init() error {
	fl.log = florist.Log.ResetNamed(fl.String())

	if err := defaults.Set(fl); err != nil {
		return fmt.Errorf("%s.init: %s", fl, err)
	}

	return nil
}

func (fl *Flower) Install(files fs.FS, finder florist.Finder) error {
	log := fl.log.Named("install")

	root, err := user.Current()
	if err != nil {
		return fmt.Errorf("%s.install: %s", fl, err)
	}

	log.Info("installing sshd configuration file")
	data := map[string]any{"Port": fl.Port}
	if err := florist.CopyTemplateFromFs(files, SshdConfig.Src, SshdConfig.Dst,
		0644, root, data, "", ""); err != nil {
		return fmt.Errorf("%s.install: %s", fl, err)
	}

	return nil
}

func (fl *Flower) Configure(files fs.FS, finder florist.Finder) error {
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
	data := secrets{
		SshHostEd25519Key:        finder.Get("SshHostEd25519Key"),
		SshHostEd25519KeyPub:     finder.Get("SshHostEd25519KeyPub"),
		SshHostEd25519KeyCertPub: finder.Get("SshHostEd25519KeyCertPub"),
	}
	if err := finder.Error(); err != nil {
		return fmt.Errorf("%s.configure: %s", fl, err)
	}

	log.Info("adding SSH host key, private")
	if err := florist.CopyTemplateFromFs(files,
		SshHostEd25519Key.Src, SshHostEd25519Key.Dst, 0400, root,
		data, "", ""); err != nil {
		return fmt.Errorf("%s.configure: %s", fl, err)
	}
	log.Info("adding SSH host key, public")
	if err := florist.CopyTemplateFromFs(files,
		SshHostEd25519KeyPub.Src, SshHostEd25519KeyPub.Dst, 0400, root,
		data, "", ""); err != nil {
		return fmt.Errorf("%s.configure: %s", fl, err)
	}
	log.Info("adding SSH host key, certificate")
	if err := florist.CopyTemplateFromFs(files,
		SshHostEd25519KeyCertPub.Src, SshHostEd25519KeyCertPub.Dst, 0400, root,
		data, "", ""); err != nil {
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
