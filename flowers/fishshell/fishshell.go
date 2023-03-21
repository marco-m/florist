// Package fishshell contains a flower to install and configure the Fish shell.
package fishshell

import (
	"fmt"
	"io/fs"
	"os/exec"
	"os/user"

	"github.com/hashicorp/go-hclog"

	"github.com/marco-m/florist"
	"github.com/marco-m/florist/pkg/apt"
)

const (
	PromptFileSrc = "prompt_hostname.fish"
	PromptFileDst = "/etc/fish/functions/prompt_hostname.fish"

	ConfigFileSrc = "config.fish"
	ConfigFileDst = "/etc/fish/config.fish"
)

var _ florist.Flower = (*Flower)(nil)

type Flower struct {
	Usernames []string
	// Using Fish as default shell breaks too many programs when they ssh :-(
	SetAsDefault bool
	log          hclog.Logger
}

func (fl *Flower) String() string {
	return "fishshell"
}

func (fl *Flower) Description() string {
	return "install and configure the Fish shell"
}

func (fl *Flower) Init() error {
	name := fmt.Sprintf("florist.flower.%s", fl)
	fl.log = florist.Log.ResetNamed(name)

	if len(fl.Usernames) == 0 {
		return fmt.Errorf("%s.new: missing usernames", name)
	}

	return nil
}

func (fl *Flower) Install(files fs.FS, finder florist.Finder) error {
	fl.log.Info("begin")
	defer fl.log.Info("end")

	fl.log.Info("Install packages")
	if err := apt.Install("fish"); err != nil {
		return err
	}

	root, err := user.Current()
	if err != nil {
		return err
	}

	fl.log.Info("Configure fish shell functions system-wide")
	// # This provides the FQDN hostname in the prompt
	if err := florist.CopyFileFromFs(files, PromptFileSrc, PromptFileDst,
		0644, root); err != nil {
		return err
	}

	fl.log.Info("Configure fish shell system-wide")
	if err := florist.CopyFileFromFs(files, ConfigFileSrc, ConfigFileDst,
		0644, root); err != nil {
		return err
	}

	found := 0
	for _, username := range fl.Usernames {
		if _, err := user.Lookup(username); err != nil {
			fl.log.Debug("user.Get", "err", err)
			continue
		}
		found++

		if fl.SetAsDefault {
			fl.log.Info("set fish shell", "user", username)
			cmd := exec.Command("chsh", "-s", "/usr/bin/fish", username)
			if err := florist.CmdRun(fl.log, cmd); err != nil {
				return err
			}
		}
	}
	if found == 0 {
		return fmt.Errorf("%s: could not find any users (%s)", fl.log.Name(),
			fl.Usernames)
	}

	return nil
}

func (fl *Flower) Configure(files fs.FS, finder florist.Finder) error {
	return nil
}
