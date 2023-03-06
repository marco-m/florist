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

const (
	AuthorizedKeysSecret = "secrets/user/authorized_keys"
)

var _ florist.Flower = (*Flower)(nil)

type Flower struct {
	fsys fs.FS
	User string
	log  hclog.Logger
}

func (fl *Flower) String() string {
	return "user"
}

func (fl *Flower) Description() string {
	return "add a user and configure SSH access"
}

func (fl *Flower) Init(fsys fs.FS) error {
	fl.fsys = fsys
	fl.log = florist.Log.Named("flower.user")

	if fl.User == "" {
		return fmt.Errorf("%s.init: missing user", fl.log.Name())
	}

	return nil
}

func (fl *Flower) Install() error {
	log := fl.log.Named("install").With("user", fl.User, "supplementary group", "docker")

	// FIXME missing SupplementaryGroups ???
	if err := florist.GroupSystemAdd("docker"); err != nil {
		return fmt.Errorf("%s: %s", log.Name(), err)
	}

	log.Info("adding user")
	user, err := florist.UserAdd(fl.User)
	if err != nil {
		return fmt.Errorf("%s: %s", log.Name(), err)
	}

	log.Info("adding SSH authorized_keys")
	if err := ssh.AddAuthorizedKeys(user, fl.fsys, AuthorizedKeysSecret); err != nil {
		return fmt.Errorf("%s: %s", log.Name(), err)
	}

	// FIXME passwordless sudo has user "ops" hardcoded
	log.Info("enabling passwordless sudo")
	if err := os.WriteFile("/etc/sudoers.d/user-ops",
		[]byte("ops ALL=(ALL) NOPASSWD: ALL\n"), 0644); err != nil {
		return fmt.Errorf("%s: %s", log.Name(), err)
	}

	return nil
}

func (fl *Flower) Configure() error {
	return nil
}