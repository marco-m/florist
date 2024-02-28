// Package nomadserver contains a flower to install a Nomad server.
package nomadserver

import (
	"embed"
	"fmt"
	"io/fs"
	"path"
	"path/filepath"

	"github.com/creasty/defaults"

	"github.com/marco-m/florist/flowers/nomad"
	"github.com/marco-m/florist/pkg/apt"
	"github.com/marco-m/florist/pkg/florist"
	"github.com/marco-m/florist/pkg/systemd"
)

//go:embed embedded
var embedded embed.FS

const (
	Name = "nomadserver"

	ConfigFile       = "embedded/nomad.server.hcl"
	UnitFile         = "embedded/nomad-server.service"
	GossipFile       = "embedded/gossip.hcl"
	CaKeyPubFile     = "embedded/CaKeyPub"
	ServerKeyFile    = "embedded/ServerKey"
	ServerKeyPubFile = "embedded/ServerKeyPub"
)

var _ florist.Flower = (*Flower)(nil)

// WARNING: Do NOT install alongside a Nomad client.
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
	DataCenter  string
	//
	NomadNumServers string
	// Apply to clients and servers.
	NomadCaKeyPub string
	// Apply only to servers.
	NomadServerGossipKey string
	NomadServerKey       string
	NomadServerKeyPub    string
}

func (fl *Flower) String() string {
	return Name
}

func (fl *Flower) Description() string {
	return "install a Nomad server (incompatible with a Nomad client)"
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

	log.Info("Install packages needed by Nomad server")
	if err := apt.Install(
		"ethtool",
	); err != nil {
		return fmt.Errorf("%s: %s", Name, err)
	}

	if err := nomad.CommonInstall(log, fl.Version, fl.Hash); err != nil {
		return fmt.Errorf("%s.install: %s", Name, err)
	}
	return nil
}

func (fl *Flower) Configure() error {
	log := florist.Log().With("flower", Name+".configure")

	dst := path.Join(nomad.CfgDir, filepath.Base(ConfigFile))
	log.Info("Install nomad server configuration file", "dst", dst)
	rendered, err := florist.TemplateFromFsWithDelims(fl.Fsys, ConfigFile, fl)
	if err != nil {
		return fmt.Errorf("%s.configure: %s", Name, err)
	}
	if err := florist.WriteFile(dst, rendered, 0640, nomad.Username); err != nil {
		return fmt.Errorf("%s.configure: %s", Name, err)
	}

	dst = path.Join(nomad.CfgDir, filepath.Base(GossipFile))
	log.Info("Install", "dst", dst)
	rendered, err = florist.TemplateFromFs(fl.Fsys, GossipFile, fl)
	if err != nil {
		return fmt.Errorf("%s.configure: %s", Name, err)
	}
	if err := florist.WriteFile(dst, rendered, 0640, nomad.Username); err != nil {
		return fmt.Errorf("%s.configure: %s", Name, err)
	}

	dst = path.Join(nomad.CfgDir, filepath.Base(CaKeyPubFile))
	log.Info("Install", "dst", dst)
	rendered, err = florist.TemplateFromFs(fl.Fsys, CaKeyPubFile, fl)
	if err != nil {
		return fmt.Errorf("%s.configure: %s", Name, err)
	}
	if err := florist.WriteFile(dst, rendered, 0640, nomad.Username); err != nil {
		return fmt.Errorf("%s.configure: %s", Name, err)
	}

	dst = path.Join(nomad.CfgDir, filepath.Base(ServerKeyFile))
	log.Info("Install", "dst", dst)
	rendered, err = florist.TemplateFromFs(fl.Fsys, ServerKeyFile, fl)
	if err != nil {
		return fmt.Errorf("%s.configure: %s", Name, err)
	}
	if err := florist.WriteFile(dst, rendered, 0640, nomad.Username); err != nil {
		return fmt.Errorf("%s.configure: %s", Name, err)
	}

	dst = path.Join(nomad.CfgDir, filepath.Base(ServerKeyPubFile))
	log.Info("Install", "dst", dst)
	rendered, err = florist.TemplateFromFs(fl.Fsys, ServerKeyPubFile, fl)
	if err != nil {
		return fmt.Errorf("%s.configure: %s", Name, err)
	}
	if err := florist.WriteFile(dst, rendered, 0640, nomad.Username); err != nil {
		return fmt.Errorf("%s.configure: %s", Name, err)
	}

	dst = path.Join("/etc/systemd/system/", filepath.Base(UnitFile))
	log.Info("Install nomad server systemd unit file", "dst", dst)
	if err := florist.CopyFileFs(fl.Fsys, UnitFile,
		dst, 0644, "root"); err != nil {
		return fmt.Errorf("%s.configure: %s", Name, err)
	}

	log.Info("Enable nomad server service to start at boot")
	if err := systemd.Enable(filepath.Base(UnitFile)); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}
	log.Info("Restart nomad server service")
	if err := systemd.Restart(filepath.Base(UnitFile)); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	return nil
}
