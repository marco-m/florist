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

type dynamicclient struct {
	Workspace string
	//
	NomadAgentCaPub         string
	GlobalClientNomadKeyPub string
	GlobalClientNomadKey    string
}

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

	fl.log.Info("Create cfg dir", "dst", nomadCfgDir)
	if err := florist.Mkdir(nomadCfgDir, "root", 0755); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	return nil
}

func (fl *ClientFlower) Configure(files fs.FS, finder florist.Finder) error {
	log := fl.log.Named("configure")

	log.Debug("loading dynamic configuration")
	data := dynamicclient{
		Workspace: finder.Get("Workspace"),
		//
		NomadAgentCaPub:         finder.Get("NomadAgentCaPub"),
		GlobalClientNomadKeyPub: finder.Get("GlobalClientNomadKeyPub"),
		GlobalClientNomadKey:    finder.Get("GlobalClientNomadKey"),
	}
	if err := finder.Error(); err != nil {
		return fmt.Errorf("%s.configure: %s", fl, err)
	}

	nomadCfgDst := path.Join(nomadCfgDir, "nomad.client.hcl")
	log.Info("Install nomad client configuration file", "dst", nomadCfgDst)
	if err := florist.CopyTemplateFs(files, "nomad.client.hcl.tpl",
		nomadCfgDst, 0640, "root", data, "<<", ">>"); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	nomadAgentCaPubDst := path.Join(nomadCfgDir, "NomadAgentCaPub")
	log.Info("Install", "dst", nomadAgentCaPubDst)
	if err := florist.CopyTemplateFs(files, "NomadAgentCaPub.tpl",
		nomadAgentCaPubDst, 0640, "root", data, "", ""); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	globalClientNomadKeyPubDst := path.Join(nomadCfgDir, "GlobalClientNomadKeyPub")
	log.Info("Install", "dst", globalClientNomadKeyPubDst)
	if err := florist.CopyTemplateFs(files, "GlobalClientNomadKeyPub.tpl",
		globalClientNomadKeyPubDst, 0640, "root", data, "", ""); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	globalClientNomadKeyDst := path.Join(nomadCfgDir, "GlobalClientNomadKey")
	log.Info("Install", "dst", globalClientNomadKeyDst)
	if err := florist.CopyTemplateFs(files, "GlobalClientNomadKey.tpl",
		globalClientNomadKeyDst, 0640, "root", data, "", ""); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	nomadUnit := path.Join("/etc/systemd/system/", "nomad-client.service")
	log.Info("Install nomad client systemd unit file", "dst", nomadUnit)
	if err := florist.CopyFileFs(files, "nomad-client.service",
		nomadUnit, 0644, "root"); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	log.Info("Enable nomad client to start at boot")
	if err := systemd.Enable("nomad-client.service"); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}
	log.Info("Restart nomad client")
	if err := systemd.Restart("nomad-client.service"); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	return nil
}
