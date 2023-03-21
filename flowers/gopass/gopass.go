// Package gopass contains a flower to install gopass.
package gopass

import (
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"time"

	"github.com/hashicorp/go-hclog"

	"github.com/marco-m/florist"
	"github.com/marco-m/florist/pkg/apt"
)

var _ florist.Flower = (*Flower)(nil)

type Flower struct {
	fsys    fs.FS
	Version string
	Hash    string
	log     hclog.Logger
}

func (fl *Flower) String() string {
	return "gopass"
}

func (fl *Flower) Description() string {
	return "install gopass"
}

func (fl *Flower) Init() error {
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

func (fl *Flower) Install(files fs.FS, finder florist.Finder) error {
	fl.log.Info("begin")
	defer fl.log.Info("end")

	fl.log.Info("Installing dependencies for gopass")
	if err := apt.Install("git", "gnupg", "rng-tools"); err != nil {
		return err
	}
	fl.log.Info("Download gopass package")
	url := fmt.Sprintf(
		"https://github.com/gopasspw/gopass/releases/download/v%s/gopass_%s_linux_amd64.deb",
		fl.Version, fl.Version)
	client := &http.Client{Timeout: 30 * time.Second}
	pkgPath, err := florist.NetFetch(client, url, florist.SHA256, fl.Hash,
		florist.WorkDir)
	if err != nil {
		return err
	}

	if err := apt.DpkgInstall(pkgPath); err != nil {
		return err
	}
	os.Remove(pkgPath)

	return nil

}

func (fl *Flower) Configure(files fs.FS, finder florist.Finder) error {
	return nil
}
