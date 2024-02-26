// Package daisy is an example flower that copies, expanding the template, one
// file at install time and one file at configure time.
package daisy

import (
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/creasty/defaults"

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
	// Base directory into which all files will be installed.
	DstDir     string `default:"/tmp/daisy"`
	PetalColor string `default:"white"`
	Perennial  bool   `default:"true"`
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

func (fl *Flower) Init() error {
	if fl.Fsys == nil {
		fl.Fsys = embedded
	}
	if err := defaults.Set(fl); err != nil {
		return fmt.Errorf("%s: %s", Name, err)
	}
	return nil
}

func (fl *Flower) Install() error {
	log := florist.Log.With("flower", Name+".install")
	userName := florist.User().Username

	dstPath := filepath.Join(fl.Inst.DstDir, InstallPlainFileDst)
	log.Debug("installing file (plain)",
		"src", InstallPlainFileSrc, "dst", dstPath)
	if err := florist.CopyFileFs(
		fl.Fsys, InstallPlainFileSrc, dstPath, 0600, userName); err != nil {
		return fmt.Errorf("%s: %s", Name, err)
	}

	dstPath = filepath.Join(fl.Inst.DstDir, InstallTmplFileDst)
	log.Debug("installing file (templated)",
		"src", InstallTmplFileSrc, "dst", dstPath)
	rendered, err := florist.TemplateFromFs(fl.Fsys, InstallTmplFileSrc, fl)
	if err != nil {
		return fmt.Errorf("%s: %s", Name, err)
	}
	username := florist.User().Username
	if err := florist.WriteFile(dstPath, rendered, 0600, username); err != nil {
		return fmt.Errorf("%s: %s", Name, err)
	}

	return nil
}

func (fl *Flower) Configure() error {
	log := florist.Log.With("flower", Name+".configure")

	dstPath := filepath.Join(fl.Inst.DstDir, ConfigTmplFileDst)
	log.Debug("installing file (templated)", "src", ConfigTmplFileSrc, "dst", dstPath)
	rendered, err := florist.TemplateFromFs(fl.Fsys, ConfigTmplFileSrc, fl)
	if err != nil {
		return fmt.Errorf("%s: %s", Name, err)
	}
	username := florist.User().Username
	if err := florist.WriteFile(dstPath, rendered, 0600, username); err != nil {
		return fmt.Errorf("%s: %s", Name, err)
	}

	return nil
}
