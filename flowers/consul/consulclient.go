package consul

import (
	"fmt"
	"io/fs"
	"os/user"
	"path"

	"github.com/hashicorp/go-hclog"

	"github.com/marco-m/florist/pkg/florist"
	"github.com/marco-m/florist/pkg/systemd"
)

var _ florist.Flower = (*ClientFlower)(nil)

// WARNING: Do NOT install alongside a Consul server.
type ClientFlower struct {
	Version string
	Hash    string
	log     hclog.Logger
}

func (fl *ClientFlower) String() string {
	return "consulclient"
}

func (fl *ClientFlower) Description() string {
	return "install a Consul client (incompatible with a Consul server)"
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
	fl.log.Info("begin")
	defer fl.log.Info("end")

	root, err := user.Current()
	if err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	fl.log.Info("Add system user 'consul'")
	_, err = florist.UserSystemAdd("consul", ConsulHome)
	if err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	if err := installConsulExe(fl.log, fl.Version, fl.Hash, root); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	consulUnit := path.Join("/etc/systemd/system/", "consul-client.service")
	fl.log.Info("Install consul client systemd unit file", "dst", consulUnit)
	if err := florist.CopyFileFromFs(files, "consul-client.service",
		consulUnit, 0644, root); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	fl.log.Info("Enable consul client service to start at boot")
	if err := systemd.Enable("consul-client.service"); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	// We do not start the service at Packer time because it is not needed and because it
	// saves state that makes reaching consensus more complicated if more than one agent.

	return nil
}

func (fl *ClientFlower) Configure(files fs.FS, finder florist.Finder) error {
	log := fl.log.Named("configure")

	log.Debug("loading dynamic configuration")
	data := dynamic{
		Workspace: finder.Get("Workspace"),
	}
	if err := finder.Error(); err != nil {
		return fmt.Errorf("%s.configure: %s", fl, err)
	}
	userConsulClient, err := user.Lookup("consul")
	if err != nil {
		return fmt.Errorf("%s.configure: %s", fl, err)
	}
	consulCfgDst := path.Join(ConsulHome, "consul.client.hcl")
	fl.log.Info("Install consul client configuration file", "dst", consulCfgDst)
	if err := florist.CopyTemplateFromFs(files,
		"consul.client.hcl.tpl", consulCfgDst, 0640, userConsulClient,
		data, "<<", ">>"); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	return nil
}
