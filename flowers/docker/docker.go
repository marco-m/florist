// packge docker contains a flower to install Docker.
package docker

import (
	"fmt"
	"time"

	"github.com/marco-m/florist"
	"github.com/marco-m/florist/pkg/apt"
)

type Flower struct {
	// Users to add to the docker supplementary group.
	Users []string
}

func (fl *Flower) Name() string {
	return "docker"
}

func (fl *Flower) Description() string {
	return "install Docker"
}

func (fl *Flower) Install() error {
	log := florist.Log.ResetNamed("florist.flower.docker")
	log.Info("begin")
	defer log.Info("end")

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
		return fmt.Errorf("%s: %s", log.Name(), err)
	}

	for _, username := range fl.Users {
		log.Info("adding user to 'docker' supplementary group", "user", username)
		florist.SupplementaryGroups(username, "docker")
	}

	return nil
}
