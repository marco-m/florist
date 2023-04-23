// Package nomad contains a flower to install a Nomad client and a flower to
// install a Nomad server.
package nomad

import (
	"fmt"
	"io/fs"
	"net/http"
	"path"
	"time"

	"github.com/hashicorp/go-hclog"

	"github.com/marco-m/florist/pkg/apt"
	"github.com/marco-m/florist/pkg/florist"
	"github.com/marco-m/florist/pkg/systemd"
)

const (
	nomadHome = "/opt/nomad"
	nomadBin  = "/usr/local/bin"
)

type dynamic struct {
	NumServers int
	Workspace  string
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

	fl.log.Info("Add system user 'nomad-server'")
	if err := florist.UserSystemAdd("nomad-server", nomadHome); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	if err := installNomadExe(fl.log, fl.Version, fl.Hash, "root"); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	return nil
}

func (fl *ServerFlower) Configure(files fs.FS, finder florist.Finder) error {
	nomadCfg := path.Join(nomadHome, "nomad.server.hcl")
	fl.log.Info("Install nomad server configuration file", "dst", nomadCfg)
	if err := florist.CopyFileFs(files, "nomad.server.hcl",
		nomadCfg, 0640, "nomad-server"); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	nomadUnit := path.Join("/etc/systemd/system/", "nomad-server.service")
	fl.log.Info("Install nomad server systemd unit file", "dst", nomadUnit)
	if err := florist.CopyFileFs(files, "nomad-server.service",
		nomadUnit, 0644, "root"); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	fl.log.Info("Enable nomad server service to start at boot")
	if err := systemd.Enable("nomad-server.service"); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	return nil
}

func installNomadExe(log hclog.Logger, version, hash, owner string) error {
	log.Info("Download Nomad package")
	url := fmt.Sprintf(
		"https://releases.hashicorp.com/nomad/%s/nomad_%s_linux_amd64.zip",
		version,
		version,
	)
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
