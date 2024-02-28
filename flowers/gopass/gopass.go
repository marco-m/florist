// Package gopass contains a flower to install gopass.
package gopass

import (
	"fmt"
	"net/http"
	url2 "net/url"
	"os"
	"time"

	"github.com/creasty/defaults"

	"github.com/marco-m/florist/pkg/apt"
	"github.com/marco-m/florist/pkg/florist"
)

const Name = "gopass"

var _ florist.Flower = (*Flower)(nil)

type Flower struct {
	Inst
	Conf
}

type Inst struct {
	Version string
	Hash    string
}

type Conf struct {
}

func (fl *Flower) String() string {
	return Name
}

func (fl *Flower) Description() string {
	return "install " + Name
}

func (fl *Flower) Embedded() []string {
	return nil
}

func (fl *Flower) Init() error {
	if err := defaults.Set(fl); err != nil {
		return fmt.Errorf("%s: %s", Name, err)
	}
	if fl.Version == "" {
		return fmt.Errorf("%s.init: %s", Name, "missing version")
	}
	if fl.Hash == "" {
		return fmt.Errorf("%s.init: %s", Name, "missing hash")
	}

	return nil
}

func (fl *Flower) Install() error {
	log := florist.Log().With("flower", Name+".install")

	log.Info("Installing dependencies for gopass")
	if err := apt.Install(
		"git",
		"gnupg",
		"rng-tools",
	); err != nil {
		return fmt.Errorf("%s: %s", Name, err)
	}
	log.Info("Download gopass package")
	uri, err := url2.JoinPath("https://github.com/gopasspw/gopass/releases/download",
		"v"+fl.Version,
		"gopass_"+fl.Version+"_linux_amd64.deb")
	if err != nil {
		return fmt.Errorf("%s: %s", Name, err)
	}
	client := &http.Client{Timeout: 30 * time.Second}
	pkgPath, err := florist.NetFetch(client, uri, florist.SHA256, fl.Hash,
		florist.WorkDir)
	if err != nil {
		return fmt.Errorf("%s: %s", Name, err)
	}

	if err := apt.DpkgInstall(pkgPath); err != nil {
		return fmt.Errorf("%s: %s", Name, err)
	}
	os.Remove(pkgPath)

	return nil

}

func (fl *Flower) Configure() error {
	return nil
}
