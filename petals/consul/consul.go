package consul

import (
	"fmt"
	"io/fs"
	"net/http"
	"os/user"
	"path"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/marco-m/florist"
	"github.com/marco-m/florist/pkg/systemd"
)

const (
	ConsulHome = "/opt/consul"
	ConsulBin  = "/usr/local/bin"
)

func ConsulServerRun(
	log hclog.Logger,
	filesFS fs.FS,
	version string,
	hash string,
) error {
	log = log.Named("petal.consulserver")
	log.Info("begin")
	defer log.Info("end")

	root, err := user.Current()
	if err != nil {
		return err
	}

	// FIXME do we need any?
	// log.Info("Install packages needed by Consul server")
	// if err := apt.Install(
	// 	"ethtool",
	// ); err != nil {
	// 	return fmt.Errorf("%s: %s", log.Name(), err)
	// }

	log.Info("Add system user 'consul-server'")
	userConsulServer, err := florist.UserSystemAdd("consul-server", ConsulHome)
	if err != nil {
		return fmt.Errorf("%s: %s", log.Name(), err)
	}

	if err := installConsulExe(log, version, hash, root); err != nil {
		return fmt.Errorf("%s: %s", log.Name(), err)
	}

	consulCfg := path.Join(ConsulHome, "consul.server.hcl")
	log.Info("Install consul server configuration file", "dst", consulCfg)
	if err := florist.CopyFromFs(filesFS, "consul/consul.server.hcl", consulCfg, 0640, userConsulServer); err != nil {
		return fmt.Errorf("%s: %s", log.Name(), err)
	}

	consulUnit := path.Join("/etc/systemd/system/", "consul-server.service")
	log.Info("Install consul server systemd unit file", "dst", consulUnit)
	if err := florist.CopyFromFs(filesFS, "consul/consul-server.service", consulUnit, 0644, root); err != nil {
		return fmt.Errorf("%s: %s", log.Name(), err)
	}

	log.Info("Enable consul server service to start at boot")
	if err := systemd.Enable("consul-server.service"); err != nil {
		return fmt.Errorf("%s: %s", log.Name(), err)
	}

	// We do not start the service at Packer time because it is not needed and because it saves state that makes reaching consensus more complicated if more than one agent.

	return nil
}

func ConsulClientRun(
	log hclog.Logger,
	filesFS fs.FS,
	version string,
	hash string,
) error {
	log = log.Named("petal.consulclient")
	log.Info("begin")
	defer log.Info("end")

	root, err := user.Current()
	if err != nil {
		return fmt.Errorf("%s: %s", log.Name(), err)
	}

	// FIXME do we need any?
	// log.Info("Install packages needed by Consul server")
	// if err := apt.Install(
	// 	"ethtool",
	// ); err != nil {
	// 	return fmt.Errorf("%s: %s", log.Name(), err)
	// }

	log.Info("Add system user 'consul-client'")
	userConsulClient, err := florist.UserSystemAdd("consul-client", ConsulHome)
	if err != nil {
		return fmt.Errorf("%s: %s", log.Name(), err)
	}

	if err := installConsulExe(log, version, hash, root); err != nil {
		return fmt.Errorf("%s: %s", log.Name(), err)
	}

	consulCfg := path.Join(ConsulHome, "consul.client.hcl")
	log.Info("Install consul client configuration file", "dst", consulCfg)
	if err := florist.CopyFromFs(filesFS, "consul/consul.client.hcl", consulCfg, 0640, userConsulClient); err != nil {
		return fmt.Errorf("%s: %s", log.Name(), err)
	}

	consulUnit := path.Join("/etc/systemd/system/", "consul-client.service")
	log.Info("Install consul client systemd unit file", "dst", consulUnit)
	if err := florist.CopyFromFs(filesFS, "consul/consul-client.service", consulUnit, 0644, root); err != nil {
		return fmt.Errorf("%s: %s", log.Name(), err)
	}

	log.Info("Enable consul client service to start at boot")
	if err := systemd.Enable("consul-client.service"); err != nil {
		return fmt.Errorf("%s: %s", log.Name(), err)
	}

	// We do not start the service at Packer time because it is not needed and because it saves state that makes reaching consensus more complicated if more than one agent.

	return nil
}

func installConsulExe(log hclog.Logger, version string, hash string, root *user.User) error {
	log.Info("Download Consul package")
	url := fmt.Sprintf("https://releases.hashicorp.com/consul/%s/consul_%s_linux_amd64.zip", version, version)
	client := &http.Client{Timeout: 30 * time.Second}
	zipPath, err := florist.NetFetch(client, url, florist.SHA256, hash, florist.WorkDir)
	if err != nil {
		return err
	}

	extracted := path.Join(florist.WorkDir, "consul")
	log.Info("Unzipping Consul package", "dst", extracted)
	if err := florist.UnzipOne(zipPath, "consul", extracted); err != nil {
		return err
	}

	exe := path.Join(ConsulBin, "consul")
	log.Info("Install consul", "dst", exe)
	if err := florist.Copy(extracted, exe, 0755, root); err != nil {
		return err
	}

	// FIXME
	// 1. it installs only for the current user
	// 2. it errors out like this if already installed
	//  /usr/local/bin/consul -autocomplete-install
	// Error executing CLI: 2 errors occurred:
	//         * already installed in /root/.bashrc
	//         * already installed at /root/.config/fish/completions/consul.fish
	//
	// maybe I can run before consul -autocomplete-uninstall ?
	// no, it is soooo stupid:
	// consul -autocomplete-uninstall
	// Error executing CLI: 2 errors occurred:
	//         * does not installed in /root/.bashrc
	//         * does not installed in /root/.config/fish

	// log.Info("Install consul shell autocomplete")
	// cmd := exec.Command(exe, "-autocomplete-install")
	// if err := cmd.Run(); err != nil {
	// 	return err
	// }

	return nil
}
