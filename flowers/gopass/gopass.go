// Package gopass contains a flower to install gopass.
package gopass

import (
	"context"
	"fmt"
	"log/slog"
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
	gdn *florist.Garden

	Version string
	Hash    string
}

type Conf struct{}

func (fl *Flower) String() string {
	return Name
}

func (fl *Flower) Description() string {
	return "install " + Name
}

func (fl *Flower) Embedded() []string {
	return nil
}

func (fl *Flower) Init(gdn *florist.Garden) error {
	fl.gdn = gdn

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
	log := slog.With("flower", Name+".install")
	apt := apt.New(fl.gdn)

	log.Info("Installing dependencies for gopass")
	if err := apt.Install(
		"git",
		"gnupg",
		"rng-tools",
	); err != nil {
		return fmt.Errorf("%s: %s", Name, err)
	}
	log.Info("Download gopass package")
	uri := fmt.Sprintf("https://github.com/gopasspw/gopass/releases/download/v%s/gopass_%s_linux_amd64.deb", fl.Version, fl.Version)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	pkgPath, err := florist.NetFetch(ctx, uri, florist.SHA256, fl.Hash,
		fl.gdn.WorkDir())
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
