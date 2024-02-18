package consul

import (
	"fmt"
	"net/http"
	"path"
	"time"

	"github.com/hashicorp/go-hclog"

	"github.com/marco-m/florist/pkg/florist"
)

const (
	ConsulHomeDir  = "/opt/consul"
	ConsulCfgDir   = "/opt/consul/config"
	ConsulBin      = "/usr/local/bin"
	ConsulUsername = "consul"
)

type Dynamic struct {
	Workspace string
}

func InstallConsulExe(log hclog.Logger, version, hash, owner string) error {
	log.Info("Download Consul package")
	url := fmt.Sprintf(
		"https://releases.hashicorp.com/consul/%s/consul_%s_linux_amd64.zip",
		version, version)
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
