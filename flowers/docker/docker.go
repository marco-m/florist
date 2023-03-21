// Package docker contains a flower to install Docker.
package docker

import (
	"fmt"
	"io/fs"
	"time"

	"github.com/hashicorp/go-hclog"

	"github.com/marco-m/florist"
	"github.com/marco-m/florist/pkg/apt"
)

var _ florist.Flower = (*Flower)(nil)

type Flower struct {
	// Users to add to the docker supplementary group.
	Users []string
	log   hclog.Logger
}

func (fl *Flower) String() string {
	return "docker"
}

func (fl *Flower) Description() string {
	return "install Docker"
}

func (fl *Flower) Init() error {
	name := fmt.Sprintf("florist.flower.%s", fl)
	fl.log = florist.Log.ResetNamed(name)
	return nil
}

func (fl *Flower) Install(files fs.FS, finder florist.Finder) error {
	fl.log.Info("begin")
	defer fl.log.Info("end")

	fl.log.Info("Add Docker upstream APT repository")
	// https://docs.docker.com/engine/install/debian/
	if err := apt.AddRepo(
		"docker",
		"https://download.docker.com/linux/debian",
		"https://download.docker.com/linux/debian/gpg",
		"1500c1f56fa9e26b9b8f42452a553675796ade0807cdce11975eb98170b3a570",
	); err != nil {
		return err
	}
	fl.log.Info("Update APT repos (needed since we just added one)")
	if err := apt.Update(0 * time.Second); err != nil {
		return err
	}

	fl.log.Info("Install packages needed by Docker upstream")
	if err := apt.Install(
		"docker-ce",
		"docker-ce-cli",
		"containerd.io",
	); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	for _, username := range fl.Users {
		fl.log.Info("adding user to 'docker' supplementary group", "user", username)
		if err := florist.SupplementaryGroups(username, "docker"); err != nil {
			return fmt.Errorf("%s: %s", fl, err)
		}
	}

	return nil
}

func (fl *Flower) Configure(files fs.FS, finder florist.Finder) error {
	return nil
}
