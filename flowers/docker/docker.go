// Package docker contains a flower to install Docker.
package docker

import (
	"fmt"
	"log/slog"
	"os/exec"

	"github.com/creasty/defaults"

	"github.com/marco-m/florist/pkg/apt"
	"github.com/marco-m/florist/pkg/florist"
	"github.com/marco-m/florist/pkg/platform"
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

type Conf struct{}

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
	const step = Name + ".install"
	errorf := makeErrorf(step)
	log := slog.With("flower", step)

	osInfo, err := platform.CollectInfo()
	if err != nil {
		return errorf("%s", err)
	}

	// I got bitten multiple times by using the Debian or Ubuntu APT packages to
	// install Docker. Let's use the recommended method instead.
	// References:
	// https://docs.docker.com/engine/install/debian/
	// https://docs.docker.com/engine/install/ubuntu/

	log.Info("Add Docker upstream APT repository")
	switch osInfo.Id {
	case "debian":
		if err := apt.AddRepo(
			"docker",
			"https://download.docker.com/linux/debian/gpg",
			"1500c1f56fa9e26b9b8f42452a553675796ade0807cdce11975eb98170b3a570",
			"https://download.docker.com/linux/debian",
		); err != nil {
			return errorf("%s", err)
		}
	case "ubuntu":
		if err := apt.AddRepo(
			"docker",
			"https://download.docker.com/linux/ubuntu/gpg",
			"1500c1f56fa9e26b9b8f42452a553675796ade0807cdce11975eb98170b3a570",
			"https://download.docker.com/linux/ubuntu",
		); err != nil {
			return errorf("%s", err)
		}
	default:
		return errorf("unsupported distribution: %s", osInfo.Id)
	}

	log.Info("Install packages needed by Docker upstream")
	if err := apt.Install(
		"docker-ce",
		"docker-ce-cli",
		"containerd.io",
		"docker-buildx-plugin",
		"docker-compose-plugin",
	); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	// Seems not needed anymore since 2023-09 ?
	//
	// 	log.Info("Enable IPv6 for the Docker daemon")
	// 	// https://github.com/docker/hub-feedback/issues/2165#issuecomment-1173017573
	// 	conf := `
	// {
	//   "registry-mirrors": [
	//     "https://registry.ipv6.docker.com"
	//   ]
	// }`
	// 	if err := os.WriteFile("/etc/docker/daemon.json", []byte(conf), 0o644); err != nil {
	// 		return fmt.Errorf("%s: %s", fl, err)
	// 	}

	log.Info("Restart the Docker daemon")
	if err := systemd.Restart("docker"); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	for _, username := range fl.Users {
		log.Info("adding user to 'docker' supplementary group", "user", username)
		if err := florist.UserMod(username, florist.UserModOpts{
			Groups: []string{"docker"},
		}); err != nil {
			return fmt.Errorf("%s: %s", fl, err)
		}
	}

	return nil
}

func (fl *Flower) Configure() error {
	const step = Name + ".configure"
	errorf := makeErrorf(step)
	log := slog.With("flower", step)

	log.Info("docker-sanity-check", "status", "running")

	if err := florist.CmdRun(log, exec.Command("docker", "run", "--rm", "hello-world")); err != nil {
		return errorf("%s", err)
	}
	if err := florist.CmdRun(log, exec.Command("docker", "system", "prune", "--force")); err != nil {
		return errorf("%s", err)
	}

	// TODO: after system prune, we can remove all images:
	// Remove:
	// docker rmi $(docker images -a -q)

	log.Info("docker-sanity-check", "status", "passed")

	return nil
}

func makeErrorf(prefix string) func(format string, a ...any) error {
	return func(format string, a ...any) error {
		return fmt.Errorf(prefix+": "+format, a...)
	}
}
