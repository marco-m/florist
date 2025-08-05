// Package fishshell contains a flower to install and configure the Fish shell.
package fishshell

import (
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"os/exec"
	"os/user"

	"github.com/creasty/defaults"

	"github.com/marco-m/florist/pkg/apt"
	"github.com/marco-m/florist/pkg/florist"
)

//go:embed embedded
var embedded embed.FS

const (
	PromptFileSrc = "embedded/prompt_hostname.fish"
	PromptFileDst = "/etc/fish/functions/prompt_hostname.fish"

	ConfigFileSrc = "embedded/config.fish"
	ConfigFileDst = "/etc/fish/config.fish"
)

const Name = "fish-shell"

var _ florist.Flower = (*Flower)(nil)

type Flower struct {
	Inst
	Conf
}

type Inst struct {
	Usernames []string
	// Using Fish as default shell breaks too many programs when they ssh :-(
	SetAsDefault bool
	Fsys         fs.FS
}

type Conf struct{}

func (fl *Flower) String() string {
	return Name
}

func (fl *Flower) Description() string {
	return "install and configure the Fish shell"
}

func (fl *Flower) Embedded() []string {
	return florist.ListFs(fl.Fsys)
}

func (fl *Flower) Init() error {
	if fl.Fsys == nil {
		fl.Fsys = embedded
	}
	if err := defaults.Set(fl); err != nil {
		return fmt.Errorf("%s: %s", Name, err)
	}
	if len(fl.Usernames) == 0 {
		return fmt.Errorf("%s.new: missing usernames", Name)
	}
	return nil
}

func (fl *Flower) Install() error {
	log := slog.With("flower", Name+".install")

	log.Info("Install packages")
	if err := apt.Install("fish"); err != nil {
		return err
	}

	log.Info("Configure fish shell functions system-wide")
	// # This provides the FQDN hostname in the prompt
	if err := florist.CopyFileFs(fl.Fsys, PromptFileSrc, PromptFileDst,
		0o644, "root"); err != nil {
		return err
	}

	log.Info("Configure fish shell system-wide")
	if err := florist.CopyFileFs(fl.Fsys, ConfigFileSrc, ConfigFileDst,
		0o644, "root"); err != nil {
		return err
	}

	found := 0
	for _, username := range fl.Usernames {
		if _, err := user.Lookup(username); err != nil {
			log.Debug("user.Get", "err", err)
			continue
		}
		found++

		if fl.SetAsDefault {
			log.Info("set fish shell", "user", username)
			cmd := exec.Command("chsh", "-s", "/usr/bin/fish", username)
			if err := florist.CmdRun(log, cmd); err != nil {
				return err
			}
		}
	}
	if found == 0 {
		return fmt.Errorf("%s: could not find any user among: %s", Name,
			fl.Usernames)
	}

	return nil
}

func (fl *Flower) Configure() error {
	log := slog.With("flower", Name+".configure")
	log.Debug("nothing-to-do")
	return nil
}
