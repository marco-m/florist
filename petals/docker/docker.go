package docker

import (
	"fmt"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/marco-m/florist"
	"github.com/marco-m/florist/pkg/apt"
)

func DockerRun(
	log hclog.Logger,
	users []string,
) error {
	log = log.Named("petal.docker")
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

	for _, username := range users {
		log.Info("adding user to docker supplementary group")
		florist.SupplementaryGroups(username, "docker")
	}

	return nil
}
