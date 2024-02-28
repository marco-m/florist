// Package consulclient contains a flower to install a Consul client.
package consulclient

import (
	"embed"
	"fmt"
	"io/fs"
	"path"
	"path/filepath"

	"github.com/creasty/defaults"

	"github.com/marco-m/florist/flowers/consul"
	"github.com/marco-m/florist/pkg/florist"
	"github.com/marco-m/florist/pkg/systemd"
)

//go:embed embedded
var embedded embed.FS

const (
	HclSrc   = "embedded/consul.client.hcl"
	UnitFile = "embedded/consul-client.service"
)

const Name = "consulclient"

var _ florist.Flower = (*Flower)(nil)

// WARNING: Do NOT install alongside a Consul server.
type Flower struct {
	Inst
	Conf
}

type Inst struct {
	Version string
	Hash    string
	Fsys    fs.FS
}

type Conf struct {
	Environment string
}

func (fl *Flower) String() string {
	return Name
}

func (fl *Flower) Description() string {
	return "install a Consul client (incompatible with a Consul server)"
}

func (fl *Flower) Embedded() []string {
	return florist.ListFs(fl.Fsys)
}

func (fl *Flower) Init() error {
	if fl.Fsys == nil {
		fl.Fsys = embedded
	}
	if err := defaults.Set(fl); err != nil {
		return fmt.Errorf("%s: %s", Name, err)
	}
	if fl.Version == "" {
		return fmt.Errorf("%s.init: %s", Name, "missing version")
	}
	if fl.Hash == "" {
		return fmt.Errorf("%s.init: %s", Name, "missing hash")
	}

	return nil
}

func (fl *Flower) Install() error {
	log := florist.Log().With("flower", Name+".install")

	if err := consul.CommonInstall(log, fl.Version, fl.Hash); err != nil {
		return fmt.Errorf("%s.install: %s", Name, err)
	}
	return nil
}

func (fl *Flower) Configure() error {
	log := florist.Log().With("flower", Name+".configure")

	dst := path.Join(consul.CfgDir, filepath.Base(HclSrc))
	log.Info("Install consul client configuration file", "dst", dst)
	rendered, err := florist.TemplateFromFsWithDelims(fl.Fsys, HclSrc, fl)
	if err != nil {
		return fmt.Errorf("%s.configure: %s", Name, err)
	}
	if err := florist.WriteFile(dst, rendered, 0640, consul.Username); err != nil {
		return fmt.Errorf("%s.configure: %s", Name, err)
	}

	dst = path.Join("/etc/systemd/system/", filepath.Base(UnitFile))
	log.Info("Install consul client systemd unit file", "dst", dst)
	if err := florist.CopyFileFs(fl.Fsys, "consul-client.service",
		dst, 0644, "root"); err != nil {
		return fmt.Errorf("%s.configure: %s", Name, err)
	}

	log.Info("Enable consul client to start at boot")
	if err := systemd.Enable(filepath.Base(UnitFile)); err != nil {
		return fmt.Errorf("%s.configure: %s", Name, err)
	}
	log.Info("Restart consul client")
	if err := systemd.Restart(filepath.Base(UnitFile)); err != nil {
		return fmt.Errorf("%s.configure: %s", Name, err)
	}

	return nil
}
