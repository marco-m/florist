package nomad

import (
	"fmt"
	"net/http"
	"path"
	"time"

	"github.com/hashicorp/go-hclog"

	"github.com/marco-m/florist/pkg/florist"
)

const (
	NomadHomeDir  = "/opt/nomad"
	NomadCfgDir   = "/opt/nomad/config"
	NomadBin      = "/usr/local/bin"
	NomadUsername = "nomad"
)

func InstallNomadExe(log hclog.Logger, version, hash, owner string) error {
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

	exe := path.Join(NomadBin, "nomad")
	log.Info("Install nomad", "dst", exe)
	if err := florist.CopyFile(extracted, exe, 0755, owner); err != nil {
		return err
	}

	// FIXME, see consul for an example
	// log.Info("Install nomad shell autocomplete")

	return nil
}
