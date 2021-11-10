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

func NomadServerRun(
	log hclog.Logger,
	filesFS fs.FS,
	version string,
	hash string,
) error {
	log = log.Named("petal.nomadserver")
	log.Info("begin")
	defer log.Info("end")

	root, err := user.Current()
	if err != nil {
		return fmt.Errorf("%s: %s", log.Name(), err)
	}

	log.Info("Install packages needed by Nomad server")
	if err := apt.Install(
		"ethtool", // Used by Nomad
	); err != nil {
		return fmt.Errorf("%s: %s", log.Name(), err)
	}

	log.Info("Add system user 'nomad-server'")
	userNomadServer, err := florist.UserSystemAdd("nomad-server", NomadServerHome)
	if err != nil {
		return fmt.Errorf("%s: %s", log.Name(), err)
	}

	if err := installNomadExe(log, version, hash, root); err != nil {
		return fmt.Errorf("%s: %s", log.Name(), err)
	}

	nomadCfg := path.Join(NomadServerHome, "nomad.server.hcl")
	log.Info("Install nomad server configuration file", "dst", nomadCfg)
	if err := florist.CopyFromFs(filesFS, "nomad/nomad.server.hcl", nomadCfg, 0640, userNomadServer); err != nil {
		return fmt.Errorf("%s: %s", log.Name(), err)
	}

	nomadUnit := path.Join("/etc/systemd/system/", "nomad-server.service")
	log.Info("Install nomad server systemd unit file", "dst", nomadUnit)
	if err := florist.CopyFromFs(filesFS, "nomad/nomad-server.service", nomadUnit, 0644, root); err != nil {
		return fmt.Errorf("%s: %s", log.Name(), err)
	}

	log.Info("Enable nomad server service to start at boot")
	if err := systemd.Enable("nomad-server.service"); err != nil {
		return fmt.Errorf("%s: %s", log.Name(), err)
	}

	// We do not start the service at Packer time because it is not needed and because it saves state that makes reaching consensus more complicated if more than one agent.

	return nil
}

func NomadClientRun(
	log hclog.Logger,
	filesFS fs.FS,
	version string,
	hash string,
) error {
	log = log.Named("petal.nomadclient")
	log.Info("begin")
	defer log.Info("end")

	root, err := user.Current()
	if err != nil {
		return fmt.Errorf("%s: %s", log.Name(), err)
	}

	log.Info("Install packages needed by Nomad client")
	if err := apt.Install(
		"apparmor",  // Needed by Nomad for the Docker driver
		"docker.io", // Needed by Nomad for the Docker driver
		"ethtool",   // Used by Nomad
	); err != nil {
		return fmt.Errorf("%s: %s", log.Name(), err)
	}

	// FIXME we add the user but we don't use it, because we need to run the nomad client as root
	log.Info("Add system user 'nomad-client'")
	_, err = florist.UserSystemAdd("nomad-client", NomadClientHome)
	if err != nil {
		return fmt.Errorf("%s: %s", log.Name(), err)
	}

	if err := installNomadExe(log, version, hash, root); err != nil {
		return fmt.Errorf("%s: %s", log.Name(), err)
	}

	nomadCfg := path.Join(NomadClientHome, "nomad.client.hcl")
	log.Info("Install nomad client configuration file", "dst", nomadCfg)
	if err := florist.CopyFromFs(filesFS, "nomad/nomad.client.hcl", nomadCfg, 0640, root); err != nil {
		return fmt.Errorf("%s: %s", log.Name(), err)
	}

	nomadUnit := path.Join("/etc/systemd/system/", "nomad-client.service")
	log.Info("Install nomad client systemd unit file", "dst", nomadUnit)
	if err := florist.CopyFromFs(filesFS, "nomad/nomad-client.service", nomadUnit, 0644, root); err != nil {
		return fmt.Errorf("%s: %s", log.Name(), err)
	}

	log.Info("Enable nomad client service to start at boot")
	if err := systemd.Enable("nomad-client.service"); err != nil {
		return fmt.Errorf("%s: %s", log.Name(), err)
	}

	// We do not start the service at Packer time because it is not needed and because it saves state that makes reaching consensus more complicated if more than one agent.

	return nil
}

func installNomadExe(log hclog.Logger, version string, hash string, root *user.User) error {
	log.Info("Download Nomad package")
	url := fmt.Sprintf("https://releases.hashicorp.com/nomad/%s/nomad_%s_linux_amd64.zip", version, version)
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

	exe := path.Join(NomadBin, "nomad")
	log.Info("Install nomad", "dst", exe)
	if err := florist.Copy(extracted, exe, 0755, root); err != nil {
		return err
	}

	// FIXME
	// log.Info("Install nomad shell autocomplete")

	return nil
}
