// Package consul contains a flower to install a Consul client and a flower to
// install a Consul server.
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

var _ florist.Flower = (*ServerFlower)(nil)

// WARNING: Do NOT install alongside a Consul client.
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
	root, err := user.Current()
	if err != nil {
		return err
	}

	fl.log.Info("Add system user 'consul'")
	userConsulServer, err := florist.UserSystemAdd("consul", ConsulHome)
	if err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	if err := installConsulExe(fl.log, fl.Version, fl.Hash, root); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	consulCfg := path.Join(ConsulHome, "consul.server.hcl")
	fl.log.Info("Install consul server configuration file", "dst", consulCfg)
	if err := florist.CopyFileFromFs(files, "consul.server.hcl",
		consulCfg, 0640, userConsulServer); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	consulUnit := path.Join("/etc/systemd/system/", "consul-server.service")
	fl.log.Info("Install consul server systemd unit file", "dst", consulUnit)
	if err := florist.CopyFileFromFs(files, "consul-server.service",
		consulUnit, 0644, root); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	fl.log.Info("Enable consul server service to start at boot")
	if err := systemd.Enable("consul-server.service"); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	// We do not start the service at Packer time because it is not needed and because it
	// saves state that makes reaching consensus more complicated if more than one agent.

	return nil
}

func (fl *ServerFlower) Configure(files fs.FS, finder florist.Finder) error {
	return nil
}

var _ florist.Flower = (*ClientFlower)(nil)

// WARNING: Do NOT install alongside a Consul server.
type ClientFlower struct {
	Version string
	Hash    string
	log     hclog.Logger
}

func (fl *ClientFlower) String() string {
	return "consulclient"
}

func (fl *ClientFlower) Description() string {
	return "install a Consul client (incompatible with a Consul server)"
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

	// FIXME do we need any?
	// fl.log.Info("Install packages needed by Consul client")
	// if err := apt.Install(
	// 	"ethtool",
	// ); err != nil {
	// 	return fmt.Errorf("%s: %s", fl, err)
	// }

	fl.log.Info("Add system user 'consul'")
	userConsulClient, err := florist.UserSystemAdd("consul", ConsulHome)
	if err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	if err := installConsulExe(fl.log, fl.Version, fl.Hash, root); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	consulCfg := path.Join(ConsulHome, "consul.client.hcl")
	fl.log.Info("Install consul client configuration file", "dst", consulCfg)
	if err := florist.CopyFileFromFs(files, "consul.client.hcl",
		consulCfg, 0640, userConsulClient); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	consulUnit := path.Join("/etc/systemd/system/", "consul-client.service")
	fl.log.Info("Install consul client systemd unit file", "dst", consulUnit)
	if err := florist.CopyFileFromFs(files, "consul-client.service",
		consulUnit, 0644, root); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	fl.log.Info("Enable consul client service to start at boot")
	if err := systemd.Enable("consul-client.service"); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	// We do not start the service at Packer time because it is not needed and because it
	// saves state that makes reaching consensus more complicated if more than one agent.

	return nil
}

func (fl *ClientFlower) Configure(files fs.FS, finder florist.Finder) error {
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
	if err := florist.CopyFile(extracted, exe, 0755, root); err != nil {
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
