// Package daisy is an example flower that copies, expanding the template, one
// file at install time and one file at configure time.
package daisy

import (
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"path/filepath"

	"github.com/marco-m/florist/pkg/florist"
)

//go:embed embedded
var embedded embed.FS

// Input paths in the source fs.
const (
	InstallPlainFileSrc = "embedded/inst1.txt"
	InstallTmplFileSrc  = "embedded/inst2.txt.tmpl"
	ConfigTmplFileSrc   = "embedded/config1.txt.tmpl"
)

// Output paths in the destination fs, relative to the customizable Flower.DstDir.
const (
	InstallPlainFileDst = "inst1.txt"
	InstallTmplFileDst  = "inst2.txt"
	ConfigTmplFileDst   = "config1.txt"
)

const Name = "daisy"

var _ florist.Flower = (*Flower)(nil)

type Flower struct {
	Inst
	Conf
}

type Inst struct {
	gdn *florist.Garden

	// Base directory into which all files will be installed.
	DstDir     string
	PetalColor string
	Perennial  *bool
	Fsys       fs.FS
}

type Conf struct {
	Environment string // dynamic setting
	GossipKey   string // Secret
}

func (fl *Flower) String() string {
	return Name
}

func (fl *Flower) Description() string {
	return "a daisy flower"
}

func (fl *Flower) Embedded() []string {
	return florist.ListFs(fl.Fsys)
}

func (fl *Flower) Init(gdn *florist.Garden) error {
	fl.gdn = gdn

	// Defaults
	if fl.Fsys == nil {
		fl.Fsys = embedded
	}
	if fl.DstDir == "" {
		fl.DstDir = filepath.Join(gdn.WorkDir(), "daisy")
	}
	if fl.PetalColor == "" {
		fl.PetalColor = "white"
	}
	if fl.Perennial == nil {
		fl.Perennial = new(true)
	}

	return nil
}

func (fl *Flower) Install() error {
	log := slog.With("flower", Name+".install")
	user := fl.gdn.User()
	group := fl.gdn.Group()

	dstPath := filepath.Join(fl.Inst.DstDir, InstallPlainFileDst)
	log.Debug("installing file (plain)",
		"src", InstallPlainFileSrc, "dst", dstPath)
	if err := florist.CopyFileFs(
		fl.Fsys, InstallPlainFileSrc, dstPath, 0o600, user,
	); err != nil {
		return fmt.Errorf("%s: %s", Name, err)
	}

	dstPath = filepath.Join(fl.Inst.DstDir, InstallTmplFileDst)
	log.Debug("installing file (templated)",
		"src", InstallTmplFileSrc, "dst", dstPath)
	rendered, err := florist.TemplateFromFs(fl.Fsys, InstallTmplFileSrc, fl)
	if err != nil {
		return fmt.Errorf("%s: %s", Name, err)
	}
	if err := florist.WriteFile(dstPath, rendered, 0o600, user, group); err != nil {
		return fmt.Errorf("%s: %s", Name, err)
	}

	return nil
}

func (fl *Flower) Configure() error {
	log := slog.With("flower", Name+".configure")
	user := fl.gdn.User()
	group := fl.gdn.Group()

	dstPath := filepath.Join(fl.Inst.DstDir, ConfigTmplFileDst)
	log.Debug("installing file (templated)", "src", ConfigTmplFileSrc, "dst", dstPath)
	rendered, err := florist.TemplateFromFs(fl.Fsys, ConfigTmplFileSrc, fl)
	if err != nil {
		return fmt.Errorf("%s: %s", Name, err)
	}
	if err := florist.Mkdir(filepath.Dir(dstPath), 0o700, user, group); err != nil {
		return fmt.Errorf("%s: %s", Name, err)
	}
	if err := florist.WriteFile(dstPath, rendered, 0o600, user, group); err != nil {
		return fmt.Errorf("%s: %s", Name, err)
	}

	return nil
}
