// Package nomad contains a flower to install a Nomad client and a flower to
// install a Nomad server.
package nomad

import (
	"fmt"
	"io/fs"
	"net/http"
	"os/user"
	"path"
	"time"

	"github.com/hashicorp/go-hclog"

	"github.com/marco-m/florist"
	"github.com/marco-m/florist/pkg/apt"
	"github.com/marco-m/florist/pkg/systemd"
)

const (
	NomadServerHome = "/opt/nomad/server"
	NomadClientHome = "/opt/nomad/client"
	NomadBin        = "/usr/local/bin"
)

var _ florist.Flower = (*ServerFlower)(nil)

// WARNING: Do NOT install alongside a Nomad client.
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

func (fl *ServerFlower) Install(files fs.FS, finder florist.Finder) error {
	fl.log.Info("begin")
	defer fl.log.Info("end")

	root, err := user.Current()
	if err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	fl.log.Info("Install packages needed by Nomad server")
	if err := apt.Install(
		"ethtool", // Used by Nomad
	); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	fl.log.Info("Add system user 'nomad-server'")
	userNomad, err := florist.UserSystemAdd("nomad-server", NomadServerHome)
	if err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	if err := installNomadExe(fl.log, fl.Version, fl.Hash, root); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	nomadCfg := path.Join(NomadServerHome, "nomad.server.hcl")
	fl.log.Info("Install nomad server configuration file", "dst", nomadCfg)
	if err := florist.CopyFileFromFs(files, "nomad.server.hcl",
		nomadCfg, 0640, userNomad); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	nomadUnit := path.Join("/etc/systemd/system/", "nomad-server.service")
	fl.log.Info("Install nomad server systemd unit file", "dst", nomadUnit)
	if err := florist.CopyFileFromFs(files, "nomad-server.service",
		nomadUnit, 0644, root); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	fl.log.Info("Enable nomad server service to start at boot")
	if err := systemd.Enable("nomad-server.service"); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	// We do not start the service at Packer time because it is not needed and
	// because it saves state that makes reaching consensus more complicated if
	// more than one agent.

	return nil
}

func (fl *ServerFlower) Configure(files fs.FS, finder florist.Finder) error {
	return nil
}

var _ florist.Flower = (*ClientFlower)(nil)

// WARNING: Do NOT install alongside a Nomad server.
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
	fl.log.Info("begin")
	defer fl.log.Info("end")

	root, err := user.Current()
	if err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	fl.log.Info("Install packages needed by Nomad client")
	if err := apt.Install(
		"apparmor", // Needed by Nomad for the Docker driver
		"ethtool",  // Used by Nomad
	); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	// FIXME we add the user but we don't use it, because we need to run the
	// nomad client as root
	fl.log.Info("Add system user 'nomad-client'")
	_, err = florist.UserSystemAdd("nomad-client", NomadClientHome)
	if err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	if err := installNomadExe(fl.log, fl.Version, fl.Hash, root); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	nomadCfg := path.Join(NomadClientHome, "nomad.client.hcl")
	fl.log.Info("Install nomad client configuration file", "dst", nomadCfg)
	if err := florist.CopyFileFromFs(files, "nomad.client.hcl",
		nomadCfg, 0640, root); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	nomadUnit := path.Join("/etc/systemd/system/", "nomad-client.service")
	fl.log.Info("Install nomad client systemd unit file", "dst", nomadUnit)
	if err := florist.CopyFileFromFs(files, "nomad-client.service",
		nomadUnit, 0644, root); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	fl.log.Info("Enable nomad client service to start at boot")
	if err := systemd.Enable("nomad-client.service"); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	// We do not start the service at Packer time because it is not needed and
	// because it saves state that makes reaching consensus more complicated if
	// more than one agent.

	return nil
}

func (fl *ClientFlower) Configure(files fs.FS, finder florist.Finder) error {
	return nil
}

func installNomadExe(
	log hclog.Logger,
	version string,
	hash string,
	root *user.User,
) error {
	log.Info("Download Nomad package")
	url := fmt.Sprintf(
		"https://releases.hashicorp.com/nomad/%s/nomad_%s_linux_amd64.zip",
		version,
		version,
	)
	client := &http.Client{Timeout: 30 * time.Second}
	zipPath, err := florist.NetFetch(client, url, florist.SHA256, hash,
		florist.WorkDir)
	if err != nil {
		return err
	}

	extracted := path.Join(florist.WorkDir, "nomad")
	log.Info("Unzipping Nomad package", "dst", extracted)
	if err := florist.UnzipOne(zipPath, "nomad", extracted); err != nil {
		return err
	}

	exe := path.Join(NomadBin, "nomad")
	log.Info("Install nomad", "dst", exe)
	if err := florist.CopyFile(extracted, exe, 0755, root); err != nil {
		return err
	}

	// FIXME
	// log.Info("Install nomad shell autocomplete")

	return nil
}
