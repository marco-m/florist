// Package consul contains a flower to install a Consul client and a flower to
// install a Consul server.
package consul

import (
	"fmt"
	"io/fs"
	"net/http"
	"path"
	"time"

	"github.com/hashicorp/go-hclog"

	"github.com/marco-m/florist/pkg/florist"
	"github.com/marco-m/florist/pkg/systemd"
)

const (
	ConsulHome = "/opt/consul"
	ConsulBin  = "/usr/local/bin"
)

type dynamic struct {
	Workspace string
}

var _ florist.Flower = (*ServerFlower)(nil)

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
	fl.log = florist.Log.ResetNamed("consulserver")

	if fl.Version == "" {
		return fmt.Errorf("consulserver.init: missing version")
	}
	if fl.Hash == "" {
		return fmt.Errorf("consulserver.init: missing hash")
	}

	return nil
}

func (fl *ServerFlower) Install(files fs.FS, finder florist.Finder) error {
	fl.log.Info("Add system user 'consul'")
	if err := florist.UserSystemAdd("consul", ConsulHome); err != nil {
		return fmt.Errorf("consulserver.install: %s", err)
	}

	if err := installConsulExe(fl.log, fl.Version, fl.Hash, "root"); err != nil {
		return fmt.Errorf("consulserver.install: %s", err)
	}

	return nil
}

func (fl *ServerFlower) Configure(files fs.FS, finder florist.Finder) error {
	log := fl.log.Named("configure")

	log.Debug("loading dynamic configuration")
	data := dynamic{
		Workspace: finder.Get("Workspace"),
	}
	if err := finder.Error(); err != nil {
		return fmt.Errorf("%s.configure: %s", fl, err)
	}

	consulCfgDst := path.Join(ConsulHome, "consul.server.hcl")
	fl.log.Info("Install consul server configuration file", "dst", consulCfgDst)
	if err := florist.CopyTemplateFs(files,
		"consul.server.hcl.tpl", consulCfgDst, 0640, "consul",
		data, "<<", ">>"); err != nil {
		return fmt.Errorf("%s.configure: %s", fl, err)
	}

	consulUnit := path.Join("/etc/systemd/system/", "consul-server.service")
	fl.log.Info("Install consul server systemd unit file", "dst", consulUnit)
	if err := florist.CopyFileFs(files, "consul-server.service",
		consulUnit, 0644, "root"); err != nil {
		return fmt.Errorf("consulserver.configure: %s", err)
	}

	fl.log.Info("Enable consul server to start at boot")
	if err := systemd.Enable("consul-server.service"); err != nil {
		return fmt.Errorf("consulserver.configure: %s", err)
	}

	fl.log.Info("Start consul server")
	if err := systemd.Restart("consul-server.service"); err != nil {
		return fmt.Errorf("consulserver.configure: %s", err)
	}

	return nil
}

func installConsulExe(log hclog.Logger, version, hash, owner string) error {
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
	if err := florist.CopyFile(extracted, exe, 0755, owner); err != nil {
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
	// sigh no:
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
