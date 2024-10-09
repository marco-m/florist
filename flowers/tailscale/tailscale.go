// Package tailscale contains a flower to install and configure Tailscale.
package tailscale

import (
	"fmt"
	"time"

	"github.com/creasty/defaults"

	"github.com/marco-m/florist/pkg/apt"
	"github.com/marco-m/florist/pkg/florist"
	"github.com/marco-m/florist/pkg/systemd"
)

const Name = "tailscale"

var _ florist.Flower = (*Flower)(nil)

type Flower struct {
	Inst
	Conf
}

type Inst struct{}

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

// Add Tailscale’s package signing key and repository:
//
// curl -fsSL https://pkgs.tailscale.com/stable/ubuntu/jammy.noarmor.gpg |
//   sudo tee /usr/share/keyrings/tailscale-archive-keyring.gpg >/dev/null
//
// curl -fsSL https://pkgs.tailscale.com/stable/ubuntu/jammy.tailscale-keyring.list |
//   sudo tee /etc/apt/sources.list.d/tailscale.list
//
// this one contains:
//
// deb [signed-by=/usr/share/keyrings/tailscale-archive-keyring.gpg] https://pkgs.tailscale.com/stable/ubuntu jammy main

func (fl *Flower) Install() error {
	log := florist.Log().With("flower", Name+".install")
	// https://tailscale.com/kb/1187/install-ubuntu-2204
	log.Info("Add Tailscale upstream APT repository")
	if err := apt.AddRepo(
		"tailscale",
		"https://pkgs.tailscale.com/stable/ubuntu/jammy.noarmor.gpg",
		"3e03dacf222698c60b8e2f990b809ca1b3e104de127767864284e6c228f1fb39",
		"https://pkgs.tailscale.com/stable/ubuntu",
	); err != nil {
		return err
	}
	log.Info("Update APT repos (needed since we just added one)")
	if err := apt.Update(0 * time.Second); err != nil {
		return err
	}

	log.Info("Enable the Tailscale daemon to start at boot")
	if err := systemd.Enable("tailscaled"); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}
	log.Info("Restart the Tailscale daemon")
	if err := systemd.Restart("tailscaled"); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	return nil
}

func (fl *Flower) Configure() error {
	return nil
}
