// Package docker contains a flower to install Docker.
package docker

import (
	"fmt"
	"os"
	"time"

	"github.com/creasty/defaults"

	"github.com/marco-m/florist/pkg/apt"
	"github.com/marco-m/florist/pkg/florist"
	"github.com/marco-m/florist/pkg/systemd"
)

const Name = "docker"

var _ florist.Flower = (*Flower)(nil)

type Flower struct {
	Inst
	Conf
}

type Inst struct {
	// Users to add to the docker supplementary group.
	Users []string
}

type Conf struct {
}

func (fl *Flower) String() string {
	return Name
}

func (fl *Flower) Description() string {
	return "install " + Name
}

func (fl *Flower) Embedded() []string {
	return nil
}

func (fl *Flower) Init() error {
	if err := defaults.Set(fl); err != nil {
		return fmt.Errorf("%s: %s", Name, err)
	}
	return nil
}

func (fl *Flower) Install() error {
	log := florist.Log.ResetNamed(Name + ".install")

	log.Info("Add Docker upstream APT repository")
	// https://docs.docker.com/engine/install/debian/
	if err := apt.AddRepo(
		"docker",
		"https://download.docker.com/linux/debian",
		"https://download.docker.com/linux/debian/gpg",
		"1500c1f56fa9e26b9b8f42452a553675796ade0807cdce11975eb98170b3a570",
	); err != nil {
		return err
	}
	log.Info("Update APT repos (needed since we just added one)")
	if err := apt.Update(0 * time.Second); err != nil {
		return err
	}

	log.Info("Install packages needed by Docker upstream")
	if err := apt.Install(
		"docker-ce",
		"docker-ce-cli",
		"containerd.io",
	); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	log.Info("Enable IPv6 for the Docker daemon")
	// https://github.com/docker/hub-feedback/issues/2165#issuecomment-1173017573
	conf := `
{
  "registry-mirrors": [
    "https://registry.ipv6.docker.com"
  ]
}`
	if err := os.WriteFile("/etc/docker/daemon.json", []byte(conf), 0644); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	log.Info("Restart the Docker daemon")
	if err := systemd.Restart("docker"); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	for _, username := range fl.Users {
		log.Info("adding user to 'docker' supplementary group", "user", username)
		if err := florist.SupplementaryGroups(username, "docker"); err != nil {
			return fmt.Errorf("%s: %s", fl, err)
		}
	}

	return nil
}

func (fl *Flower) Configure() error {
	return nil
}
