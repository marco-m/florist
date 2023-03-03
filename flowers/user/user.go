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

var _ florist.Flower = (*Flower)(nil)

type Flower struct {
	fsys      fs.FS
	SecretsFS fs.FS
	User      string
	log       hclog.Logger
}

func (fl *Flower) String() string {
	return "user"
}

func (fl *Flower) Description() string {
	return "add a user and configure SSH access"
}

func (fl *Flower) Init(fsys fs.FS) error {
	fl.fsys = fsys
	name := fmt.Sprintf("florist.flower.%s", fl)
	fl.log = florist.Log.ResetNamed(name)

	if fl.User == "" {
		return fmt.Errorf("%s.new: missing user", name)
	}

	return nil
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
	if err := ssh.AddAuthorizedKeys(user, fl.SecretsFS,
		"secrets/user/authorized_keys"); err != nil {
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

func (fl *Flower) Configure() error {
	return nil
}
