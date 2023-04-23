package nomad

import (
	"fmt"
	"io/fs"
	"path"

	"github.com/hashicorp/go-hclog"

	"github.com/marco-m/florist/pkg/apt"
	"github.com/marco-m/florist/pkg/florist"
	"github.com/marco-m/florist/pkg/systemd"
)

var _ florist.Flower = (*ClientFlower)(nil)

type ClientFlower struct {
	Version string
	Hash    string
	log     hclog.Logger
}

func (fl *ClientFlower) String() string {
	return "nomadclient"
}

func (fl *ClientFlower) Description() string {
	return "install a Nomad client (incompatible with a Nomad server)"
}

func (fl *ClientFlower) Init() error {
	name := fmt.Sprintf("florist.flower.%s", fl)
	fl.log = florist.Log.ResetNamed(name)

	if fl.Version == "" {
		return fmt.Errorf("%s.new: missing version", name)
	}
	if fl.Hash == "" {
		return fmt.Errorf("%s.new: missing hash", name)
	}

	return nil
}

func (fl *ClientFlower) Install(files fs.FS, finder florist.Finder) error {
	fl.log.Info("Install packages needed by Nomad client")
	if err := apt.Install(
		"apparmor", // Needed by Nomad for the Docker driver
		"ethtool",  // Used by Nomad
	); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	fl.log.Info("The nomad client (contrary to the server) must run as root, so not adding any dedicated user")

	if err := installNomadExe(fl.log, fl.Version, fl.Hash, "root"); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	return nil
}

func (fl *ClientFlower) Configure(files fs.FS, finder florist.Finder) error {
	nomadCfg := path.Join(nomadHome, "nomad.client.hcl")
	fl.log.Info("Install nomad client configuration file", "dst", nomadCfg)
	if err := florist.CopyFileFs(files, "nomad.client.hcl",
		nomadCfg, 0640, "root"); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	nomadUnit := path.Join("/etc/systemd/system/", "nomad-client.service")
	fl.log.Info("Install nomad client systemd unit file", "dst", nomadUnit)
	if err := florist.CopyFileFs(files, "nomad-client.service",
		nomadUnit, 0644, "root"); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	fl.log.Info("Enable nomad client service to start at boot")
	if err := systemd.Enable("nomad-client.service"); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	return nil
}
