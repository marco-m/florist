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

type Options struct {
	FilesFS   fs.FS
	Usernames []string
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
	if len(fl.Usernames) == 0 {
		return nil, fmt.Errorf("%s.new: missing usernames", name)
	}

	return &fl, nil
}

func (fl Flower) String() string {
	return "fishshell"
}

func (fl Flower) Description() string {
	return "install and configure the Fish shell"
}

func (fl *Flower) Install() error {
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
	if err := florist.CopyFromFs(fl.FilesFS, "fish/prompt_hostname.fish",
		"/etc/fish/functions/prompt_hostname.fish", 0644, root); err != nil {
		return err
	}

	fl.log.Info("Configure fish shell system-wide")
	if err := florist.CopyFromFs(fl.FilesFS, "fish/config.fish",
		"/etc/fish/config.fish", 0644, root); err != nil {
		return err
	}

	found := 0
	for _, username := range fl.Usernames {
		if _, err := user.Lookup(username); err != nil {
			fl.log.Debug("user.Lookup", "err", err)
			continue
		}
		found++

		fl.log.Info("set fish shell", "user", username)
		cmd := exec.Command("chsh", "-s", "/usr/bin/fish", username)
		if err := florist.LogRun(fl.log, cmd); err != nil {
			return err
		}
	}
	if found == 0 {
		return fmt.Errorf("%s: could not find any users (%s)", fl.log.Name(), fl.Usernames)
	}

	return nil
}
