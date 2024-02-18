// Package consul contains a flower to install a Consul client and a flower to
// install a Consul server.
package consulserver

import (
	"fmt"
	"io/fs"
	"path"

	"github.com/hashicorp/go-hclog"

	"github.com/marco-m/florist/flowers/consul"
	"github.com/marco-m/florist/pkg/florist"
	"github.com/marco-m/florist/pkg/systemd"
)

var _ florist.Flower = (*ServerFlower)(nil)

type ServerFlower struct {
	Version string
	Hash    string
	log     hclog.Logger
}

func (fl *ServerFlower) String() string {
	return "consulserver"
}

func (fl *ServerFlower) Description() string {
	return "install a Consul server (incompatible with a Consul client)"
}

func (fl *ServerFlower) Init() error {
	fl.log = florist.Log.ResetNamed("consulserver")

	if fl.Version == "" {
		return fmt.Errorf("consulserver.init: missing version")
	}
	if fl.Hash == "" {
		return fmt.Errorf("consulserver.init: missing hash")
	}

	return nil
}

func (fl *ServerFlower) Install(files fs.FS, finder florist.Finder) error {
	log := fl.log.Named("install")

	log.Info("Add system user", "user", consul.ConsulUsername)
	if err := florist.UserSystemAdd(consul.ConsulUsername, consul.ConsulHomeDir); err != nil {
		return fmt.Errorf("consulserver.install: %s", err)
	}

	if err := consul.InstallConsulExe(log, fl.Version, fl.Hash, "root"); err != nil {
		return fmt.Errorf("consulserver.install: %s", err)
	}

	log.Info("Create cfg dir", "dst", consul.ConsulCfgDir)
	if err := florist.Mkdir(consul.ConsulCfgDir, consul.ConsulUsername, 0755); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	return nil
}

func (fl *ServerFlower) Configure(files fs.FS, finder florist.Finder) error {
	log := fl.log.Named("configure")

	log.Debug("loading dynamic configuration")
	data := consul.Dynamic{
		Workspace: finder.Get("Workspace"),
	}
	if err := finder.Error(); err != nil {
		return fmt.Errorf("%s.configure: %s", fl, err)
	}

	consulCfgDst := path.Join(consul.ConsulCfgDir, "consul.server.hcl")
	log.Info("Install consul server configuration file", "dst", consulCfgDst)
	if err := florist.CopyTemplateFs(files, "consul.server.hcl.tpl",
		consulCfgDst, 0640, consul.ConsulUsername, data, "<<", ">>"); err != nil {
		return fmt.Errorf("%s.configure: %s", fl, err)
	}

	consulUnitDst := path.Join("/etc/systemd/system/", "consul-server.service")
	log.Info("Install consul server systemd unit file", "dst", consulUnitDst)
	if err := florist.CopyFileFs(files, "consul-server.service",
		consulUnitDst, 0644, "root"); err != nil {
		return fmt.Errorf("consulserver.configure: %s", err)
	}

	log.Info("Enable consul server to start at boot")
	if err := systemd.Enable("consul-server.service"); err != nil {
		return fmt.Errorf("consulserver.configure: %s", err)
	}
	log.Info("Restart consul server")
	if err := systemd.Restart("consul-server.service"); err != nil {
		return fmt.Errorf("consulserver.configure: %s", err)
	}

	return nil
}
