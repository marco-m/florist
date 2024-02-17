// Package nomad should NOT be imported by client code.
// Instead, use packages nomadclient and nomadserver.
package nomad

import (
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/hashicorp/go-hclog"

	"github.com/marco-m/florist/pkg/florist"
)

const (
	HomeDir  = "/opt/nomad"
	CfgDir   = "/opt/nomad/config"
	BinDir   = "/usr/local/bin"
	Username = "nomad"
)

// CommonInstall performs the installation steps common to the client and the server.
func CommonInstall(log hclog.Logger, version string, hash string) error {
	// The nomad client (contrary to the server) must run as root, so we leave
	// to the consulserver flower the care of adding a dedicated user.
	// Contrast with [consul.CommonInstall].

	log.Info("Add system user", "user", Username)
	if err := florist.UserSystemAdd(Username, HomeDir); err != nil {
		return err
	}

	if err := installNomadExe(log, version, hash); err != nil {
		return err
	}

	log.Info("Create cfg dir", "dst", CfgDir)
	if err := florist.Mkdir(CfgDir, Username, 0755); err != nil {
		return err
	}

	return nil
}

func installNomadExe(log hclog.Logger, version string, hash string) error {
	log.Info("Download Nomad package")
	uri, err := url.JoinPath("https://releases.hashicorp.com/nomad",
		version, "nomad_"+version+"_linux_amd64.zip")
	if err != nil {
		return err
	}
	client := &http.Client{Timeout: 30 * time.Second}
	zipPath, err := florist.NetFetch(client, uri, florist.SHA256, hash, florist.WorkDir)
	if err != nil {
		return err
	}

	extracted := path.Join(florist.WorkDir, "nomad")
	log.Info("Unzipping Nomad package", "dst", extracted)
	if err := florist.UnzipOne(zipPath, "nomad", extracted); err != nil {
		return err
	}

	exeDst := path.Join(BinDir, "nomad")
	log.Info("Install nomad executbale", "dst", exeDst)
	if err := florist.CopyFile(extracted, exeDst, 0755, "root"); err != nil {
		return err
	}

	// FIXME, see consul for an example
	// log.Info("Install nomad shell autocomplete")

	return nil
}
