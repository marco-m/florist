// Package nomadclient contains a flower to install a Nomad client.
package nomadclient

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
	Name = "nomadclient"

	ConfigFile       = "embedded/nomad.client.hcl"
	UnitFile         = "embedded/nomad-client.service"
	CaKeyPubFile     = "embedded/CaKeyPub"
	ClientKeyFile    = "embedded/ClientKey"
	ClientKeyPubFile = "embedded/ClientKeyPub"
)

var _ florist.Flower = (*Flower)(nil)

// WARNING: Do NOT install alongside a Nomad server.
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
	//

	// Apply to client and servers.
	NomadCaPub string
	// Apply only to clients.
	NomadClientKey    string
	NomadClientKeyPub string
}

func (fl *Flower) String() string {
	return Name
}

func (fl *Flower) Description() string {
	return "install a Nomad client (incompatible with a Nomad server)"
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
	log := florist.Log.With("flower", Name+".install")

	log.Info("Install packages needed by Nomad client")
	if err := apt.Install(
		"apparmor",
		"ethtool",
		"dmidecode", // https://developer.hashicorp.com/nomad/docs/concepts/cpu#virtual-cpu-fingerprinting
	); err != nil {
		return fmt.Errorf("%s: %s", Name, err)
	}

	//log.Info("The nomad client (contrary to the server) must run as root, so not adding any dedicated user")

	if err := nomad.CommonInstall(log, fl.Version, fl.Hash); err != nil {
		return fmt.Errorf("%s.install: %s", Name, err)
	}
	return nil
}

func (fl *Flower) Configure() error {
	log := florist.Log.With("flower", Name+".configure")

	dst := path.Join(nomad.CfgDir, filepath.Base(ConfigFile))
	log.Info("Install nomad client configuration file", "dst", dst)
	rendered, err := florist.TemplateFromFsWithDelims(fl.Fsys, ConfigFile, fl)
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

	dst = path.Join(nomad.CfgDir, filepath.Base(ClientKeyFile))
	log.Info("Install", "dst", dst)
	rendered, err = florist.TemplateFromFs(fl.Fsys, ClientKeyFile, fl)
	if err != nil {
		return fmt.Errorf("%s.configure: %s", Name, err)
	}
	if err := florist.WriteFile(dst, rendered, 0640, nomad.Username); err != nil {
		return fmt.Errorf("%s.configure: %s", Name, err)
	}

	dst = path.Join(nomad.CfgDir, filepath.Base(ClientKeyPubFile))
	log.Info("Install", "dst", dst)
	rendered, err = florist.TemplateFromFs(fl.Fsys, ClientKeyPubFile, fl)
	if err != nil {
		return fmt.Errorf("%s.configure: %s", Name, err)
	}
	if err := florist.WriteFile(dst, rendered, 0640, nomad.Username); err != nil {
		return fmt.Errorf("%s.configure: %s", Name, err)
	}

	dst = path.Join("/etc/systemd/system/", filepath.Base(UnitFile))
	log.Info("Install nomad client systemd unit file", "dst", dst)
	if err := florist.CopyFileFs(fl.Fsys, UnitFile,
		dst, 0644, "root"); err != nil {
		return fmt.Errorf("%s.configure: %s", Name, err)
	}

	log.Info("Enable nomad client service to start at boot")
	if err := systemd.Enable(filepath.Base(UnitFile)); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}
	log.Info("Restart nomad client service")
	if err := systemd.Restart(filepath.Base(UnitFile)); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	return nil
}
