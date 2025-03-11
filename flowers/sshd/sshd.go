// Package sshd contains a flower to configure the OpenSSH sshd server.
//
// To generate host keys, see README.
//
// Since Ubuntu does funky modifications to the OpenSSH default file, we use as
// reference the OpenSSH default file, that we keep here to ease diffing with a
// new version.
//
// Workflow:
//   - curl --output embedded/sshd_config.upstream --fail --no-progress-meter --location https://github.com/openssh/openssh-portable/raw/master/sshd_config
//   - first time:
//   - cp embedded/sshd_config.upstream embedded/sshd_config
//   - edit embedded/sshd_config to taste and to add template expansion
//   - subsequent times:
//   - git diff embedded/sshd_config.upstream
//   - if any changes in the diff, decide what to incorporate in embedded/sshd_config
package sshd

import (
	"embed"
	"fmt"
	"io/fs"
	"os/exec"
	"strings"

	"github.com/creasty/defaults"

	"github.com/marco-m/florist/pkg/florist"
	"github.com/marco-m/florist/pkg/systemd"
)

//go:embed embedded
var embedded embed.FS

const (
	SshdConfigSrc = "embedded/sshd_config"
	SshdConfigDst = "/etc/ssh/sshd_config"

	SshHostEd25519KeyDst        = "/etc/ssh/ssh_host_ed25519_key"
	SshHostEd25519KeyPubDst     = "/etc/ssh/ssh_host_ed25519_key.pub"
	SshHostEd25519KeyCertPubDst = "/etc/ssh/ssh_host_ed25519_key-cert.pub"
)

const Name = "sshd"

var _ florist.Flower = (*Flower)(nil)

type Flower struct {
	Inst
	Conf
}

type Inst struct {
	Fsys fs.FS
}

type Conf struct {
	Port int `default:"22"`

	SshHostEd25519Key        string
	SshHostEd25519KeyPub     string
	SshHostEd25519KeyCertPub string
}

func (fl *Flower) String() string {
	return "sshd"
}

func (fl *Flower) Description() string {
	return "configure an already-existing sshd server"
}

func (fl *Flower) Embedded() []string {
	return florist.ListFs(fl.Fsys)
}

func (fl *Flower) Init() error {
	if fl.Fsys == nil {
		fl.Fsys = embedded
	}
	if err := defaults.Set(fl); err != nil {
		return fmt.Errorf("%s.init: %s", fl, err)
	}

	return nil
}

func (fl *Flower) Install() error {
	log := florist.Log().With("flower", Name+".install")
	log.Debug("nothing-to-do")
	return nil
}

func (fl *Flower) Configure() error {
	log := florist.Log().With("flower", Name+".configure")

	log.Info("installing sshd configuration file",
		"src", SshdConfigSrc, "dst", SshdConfigDst)
	rendered, err := florist.TemplateFromFs(fl.Fsys, SshdConfigSrc, fl)
	if err != nil {
		return fmt.Errorf("%s: %s", Name, err)
	}
	if err := florist.WriteFile(SshdConfigDst, rendered,
		0o644, "root", "root"); err != nil {
		return fmt.Errorf("%s: %s", Name, err)
	}

	log.Info("adding SSH host key, private")
	data := strings.TrimSpace(fl.SshHostEd25519Key) + "\n"
	if err := florist.WriteFile(SshHostEd25519KeyDst, data,
		0o600, "root", "root"); err != nil {
		return fmt.Errorf("%s: %s", Name, err)
	}

	log.Info("adding SSH host key, public")
	data = fl.SshHostEd25519KeyPub
	if err := florist.WriteFile(SshHostEd25519KeyPubDst, data,
		0o644, "root", "root"); err != nil {
		return fmt.Errorf("%s: %s", Name, err)
	}

	log.Info("adding SSH host key, certificate")
	data = fl.SshHostEd25519KeyCertPub
	if err := florist.WriteFile(SshHostEd25519KeyCertPubDst, data,
		0o644, "root", "root"); err != nil {
		return fmt.Errorf("%s: %s", Name, err)
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
