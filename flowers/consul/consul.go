// Package consul should NOT be imported by client code.
// Instead, use packages consulclient and consulserver.
package consul

import (
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/marco-m/florist/pkg/florist"
)

const (
	HomeDir  = "/opt/consul"
	CfgDir   = "/opt/consul/config"
	BinDir   = "/usr/local/bin"
	Username = "consul"
)

// CommonInstall performs the install steps common to the client and the server.
func CommonInstall(log *slog.Logger, version string, hash string) error {
	log.Info("Add system user", "user", Username)
	if err := florist.UserAdd(Username, &florist.UserAddArgs{
		System:  true,
		HomeDir: HomeDir,
	}); err != nil {
		return err
	}

	if err := installConsulExe(log, version, hash); err != nil {
		return err
	}

	log.Info("Create cfg dir", "dst", CfgDir)
	if err := os.MkdirAll(CfgDir, 0o755); err != nil {
		return err
	}

	return nil
}

func installConsulExe(log *slog.Logger, version string, hash string) error {
	log.Info("Download Consul package")
	uri, err := url.JoinPath("https://releases.hashicorp.com/consul",
		version, "consul_"+version+"_linux_amd64.zip")
	if err != nil {
		return err
	}
	client := &http.Client{Timeout: 30 * time.Second}
	zipPath, err := florist.NetFetch(client, uri, florist.SHA256, hash, florist.WorkDir)
	if err != nil {
		return err
	}

	extracted := path.Join(florist.WorkDir, "consul")
	log.Info("Unzipping Consul package", "dst", extracted)
	if err := florist.UnzipOne(zipPath, "consul", extracted); err != nil {
		return err
	}

	exeDst := path.Join(BinDir, "consul")
	log.Info("Install consul executable", "dst", exeDst)
	if err := florist.CopyFile(extracted, exeDst, 0o755, "root"); err != nil {
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
	// if err := cmd.Install(); err != nil {
	// 	return err
	// }

	return nil
}
