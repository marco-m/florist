// Package user contains a flower to add a user and configure SSH access.
package user

import (
	"fmt"
	"io/fs"
	"os/user"

	"github.com/hashicorp/go-hclog"

	"github.com/marco-m/florist/pkg/florist"
	"github.com/marco-m/florist/pkg/ssh"
)

const (
	AuthorizedKeys = "authorized_keys"
)

var _ florist.Flower = (*Flower)(nil)

type Flower struct {
	User string
	log  hclog.Logger
}

func (fl *Flower) String() string {
	return "user"
}

func (fl *Flower) Description() string {
	return "add a user and configure SSH access"
}

func (fl *Flower) Init() error {
	fl.log = florist.Log.Named("flower.user")

	if fl.User == "" {
		return fmt.Errorf("%s.init: missing user", fl.log.Name())
	}

	return nil
}

func (fl *Flower) Install(files fs.FS, finder florist.Finder) error {
	log := fl.log.Named("install").With("user", fl.User, "supplementary group", "docker")

	// FIXME missing SupplementaryGroups ???
	if err := florist.GroupSystemAdd("docker"); err != nil {
		return fmt.Errorf("%s: %s", log.Name(), err)
	}

	log.Info("adding user")
	_, err := florist.UserAdd(fl.User)
	if err != nil {
		return fmt.Errorf("%s: %s", log.Name(), err)
	}

	sudoersDst := "/etc/sudoers.d/user-" + fl.User
	log.Info("installing sudoers", "dst", sudoersDst)
	root, _ := user.Current()
	if err := florist.Mkdir("/etc/sudoers.d", root, 0755); err != nil {
		return fmt.Errorf("%s: %s", log.Name(), err)
	}
	if err := florist.CopyTemplateFromFs(files,
		"sudoers.tpl", sudoersDst, 0644, root,
		fl, "", ""); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	return nil
}

func (fl *Flower) Configure(files fs.FS, finder florist.Finder) error {
	log := fl.log.Named("configure")

	log.Debug("loading secrets")
	content := finder.Get(AuthorizedKeys)
	if err := finder.Error(); err != nil {
		return fmt.Errorf("%s: %s", log.Name(), err)
	}

	userinfo, err := user.Lookup(fl.User)
	if err != nil {
		return fmt.Errorf("user: lookup: %s", err)
	}

	log.Info("adding SSH authorized_keys")
	if err := ssh.AddAuthorizedKeys(userinfo, content); err != nil {
		return fmt.Errorf("%s: %s", log.Name(), err)
	}

	return nil
}
