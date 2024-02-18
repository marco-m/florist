package consul

import (
	"fmt"
	"io/fs"
	"path"

	"github.com/hashicorp/go-hclog"

	"github.com/marco-m/florist/flowers/consul"
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
	fl.log.Info("Add system user", "user", consul.ConsulUsername)
	if err := florist.UserSystemAdd(consul.ConsulUsername, consul.ConsulHomeDir); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	if err := consul.InstallConsulExe(fl.log, fl.Version, fl.Hash, "root"); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	fl.log.Info("Create cfg dir", "dst", consul.ConsulCfgDir)
	if err := florist.Mkdir(consul.ConsulCfgDir, consul.ConsulUsername, 0755); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	return nil
}

func (fl *ClientFlower) Configure(files fs.FS, finder florist.Finder) error {
	log := fl.log.Named("configure")

	log.Debug("loading dynamic configuration")
	data := consul.Dynamic{
		Workspace: finder.Get("Workspace"),
	}
	if err := finder.Error(); err != nil {
		return fmt.Errorf("%s.configure: %s", fl, err)
	}

	consulCfgDst := path.Join(consul.ConsulCfgDir, "consul.client.hcl")
	fl.log.Info("Install consul client configuration file", "dst", consulCfgDst)
	if err := florist.CopyTemplateFs(files, "consul.client.hcl.tpl",
		consulCfgDst, 0640, consul.ConsulUsername, data, "<<", ">>"); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	consulUnit := path.Join("/etc/systemd/system/", "consul-client.service")
	fl.log.Info("Install consul client systemd unit file", "dst", consulUnit)
	if err := florist.CopyFileFs(files, "consul-client.service",
		consulUnit, 0644, "root"); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	fl.log.Info("Enable consul client to start at boot")
	if err := systemd.Enable("consul-client.service"); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}
	fl.log.Info("Restart consul client")
	if err := systemd.Restart("consul-client.service"); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	return nil
}
