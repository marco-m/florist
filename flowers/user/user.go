// Package user contains a flower to add a user and configure SSH access.
package user

import (
	"fmt"
	"io/fs"
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/marco-m/florist"
	"github.com/marco-m/florist/pkg/ssh"
)

type Options struct {
	FilesFS fs.FS
	User    string
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
	if fl.User == "" {
		return nil, fmt.Errorf("%s.new: missing user", name)
	}

	return &fl, nil
}

func (fl Flower) String() string {
	return "user"
}

func (fl Flower) Description() string {
	return "add a user and configure SSH access"
}

func (fl *Flower) Install() error {
	fl.log.Info("begin")
	defer fl.log.Info("end")

	// FIXME missing SupplementaryGroups ???
	fl.log = fl.log.With("user", fl.User, "supplementary group", "docker")
	if err := florist.GroupSystemAdd("docker"); err != nil {
		return err
	}

	fl.log.Info("Add user")
	user, err := florist.UserAdd(fl.User)
	if err != nil {
		return err
	}

	fl.log.Info("Add SSH authorized_keys")
	if err := ssh.AddAuthorizedKeys(user, fl.FilesFS,
		"secrets/authorized_keys"); err != nil {
		return err
	}

	// FIXME passwordless sudo has user "ops" hardcoded
	fl.log.Info("Enable passwordless sudo")
	if err := os.WriteFile("/etc/sudoers.d/user-ops",
		[]byte("ops ALL=(ALL) NOPASSWD: ALL\n"), 0644); err != nil {
		return err
	}

	return nil
}