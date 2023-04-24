// Package nomad contains a flower to install a Nomad client and a flower to
// install a Nomad server.
package nomad

import (
	"fmt"
	"io/fs"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/hashicorp/go-hclog"

	"github.com/marco-m/florist/pkg/apt"
	"github.com/marco-m/florist/pkg/florist"
	"github.com/marco-m/florist/pkg/systemd"
)

const (
	nomadHomeDir  = "/opt/nomad"
	nomadCfgDir   = "/opt/nomad/config"
	nomadBin      = "/usr/local/bin"
	nomadUsername = "nomad"
)

type dynamicserver struct {
	NumServers string
	Workspace  string
	//
	GossipKey               string
	NomadAgentCaPub         string
	GlobalServerNomadKeyPub string
	GlobalServerNomadKey    string
}

var _ florist.Flower = (*ServerFlower)(nil)

type ServerFlower struct {
	Version string
	Hash    string
	log     hclog.Logger
}

func (fl *ServerFlower) String() string {
	return "nomadserver"
}

func (fl *ServerFlower) Description() string {
	return "install a Nomad server (incompatible with a Nomad client)"
}

func (fl *ServerFlower) Init() error {
	fl.log = florist.Log.ResetNamed("nomadserver")

	if fl.Version == "" {
		return fmt.Errorf("nomadserver.init: missing version")
	}
	if fl.Hash == "" {
		return fmt.Errorf("nomadserver.init: missing hash")
	}

	return nil
}

func (fl *ServerFlower) Install(files fs.FS, finder florist.Finder) error {
	fl.log.Info("Install packages needed by Nomad server")
	if err := apt.Install(
		"ethtool", // Used by Nomad
	); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	fl.log.Info("Add system user", "user", nomadUsername)
	if err := florist.UserSystemAdd(nomadUsername, nomadHomeDir); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	if err := installNomadExe(fl.log, fl.Version, fl.Hash, "root"); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	fl.log.Info("Create cfg dir", "dst", nomadCfgDir)
	if err := florist.Mkdir(nomadCfgDir, nomadUsername, 0755); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	return nil
}

func (fl *ServerFlower) Configure(files fs.FS, finder florist.Finder) error {
	log := fl.log.Named("configure")

	log.Debug("loading dynamic configuration")
	data := dynamicserver{
		NumServers: finder.Get("NumServers"),
		Workspace:  finder.Get("Workspace"),
		//
		GossipKey:               strings.TrimSpace(finder.Get("GossipKey")),
		NomadAgentCaPub:         finder.Get("NomadAgentCaPub"),
		GlobalServerNomadKeyPub: finder.Get("GlobalServerNomadKeyPub"),
		GlobalServerNomadKey:    finder.Get("GlobalServerNomadKey"),
	}
	if err := finder.Error(); err != nil {
		return fmt.Errorf("%s.configure: %s", fl, err)
	}

	nomadCfgDst := path.Join(nomadCfgDir, "nomad.server.hcl")
	log.Info("Install nomad server configuration file", "dst", nomadCfgDst)
	if err := florist.CopyTemplateFs(files, "nomad.server.hcl.tpl",
		nomadCfgDst, 0640, nomadUsername, data, "<<", ">>"); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	nomadGossipDst := path.Join(nomadCfgDir, "gossip.hcl")
	log.Info("Install", "dst", nomadGossipDst)
	if err := florist.CopyTemplateFs(files, "gossip.hcl.tpl",
		nomadGossipDst, 0640, nomadUsername, data, "<<", ">>"); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	nomadAgentCaPubDst := path.Join(nomadCfgDir, "NomadAgentCaPub")
	log.Info("Install", "dst", nomadAgentCaPubDst)
	if err := florist.CopyTemplateFs(files, "NomadAgentCaPub.tpl",
		nomadAgentCaPubDst, 0640, nomadUsername, data, "", ""); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	globalServerNomadKeyPubDst := path.Join(nomadCfgDir, "GlobalServerNomadKeyPub")
	log.Info("Install", "dst", globalServerNomadKeyPubDst)
	if err := florist.CopyTemplateFs(files, "GlobalServerNomadKeyPub.tpl",
		globalServerNomadKeyPubDst, 0640, nomadUsername, data, "", ""); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	globalServerNomadKeyDst := path.Join(nomadCfgDir, "GlobalServerNomadKey")
	log.Info("Install", "dst", globalServerNomadKeyDst)
	if err := florist.CopyTemplateFs(files, "GlobalServerNomadKey.tpl",
		globalServerNomadKeyDst, 0640, nomadUsername, data, "", ""); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	nomadUnitDst := path.Join("/etc/systemd/system/", "nomad-server.service")
	log.Info("Install nomad server systemd unit file", "dst", nomadUnitDst)
	if err := florist.CopyFileFs(files, "nomad-server.service",
		nomadUnitDst, 0644, "root"); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	log.Info("Enable nomad server service to start at boot")
	if err := systemd.Enable("nomad-server.service"); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}
	log.Info("Restart nomad server service")
	if err := systemd.Restart("nomad-server.service"); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	return nil
}

func installNomadExe(log hclog.Logger, version, hash, owner string) error {
	log.Info("Download Nomad package")
	url := fmt.Sprintf(
		"https://releases.hashicorp.com/nomad/%s/nomad_%s_linux_amd64.zip",
		version, version)
	client := &http.Client{Timeout: 30 * time.Second}
	zipPath, err := florist.NetFetch(client, url, florist.SHA256, hash, florist.WorkDir)
	if err != nil {
		return err
	}

	extracted := path.Join(florist.WorkDir, "nomad")
	log.Info("Unzipping Nomad package", "dst", extracted)
	if err := florist.UnzipOne(zipPath, "nomad", extracted); err != nil {
		return err
	}

	exe := path.Join(nomadBin, "nomad")
	log.Info("Install nomad", "dst", exe)
	if err := florist.CopyFile(extracted, exe, 0755, owner); err != nil {
		return err
	}

	// FIXME, see consul for an example
	// log.Info("Install nomad shell autocomplete")

	return nil
}
