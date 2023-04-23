// Package daisy is an example flower that copies, expanding the template, one
// file at install time and one file at configure time.
package daisy

import (
	"fmt"
	"io/fs"
	"os/user"
	"path/filepath"

	"github.com/creasty/defaults"
	"github.com/hashicorp/go-hclog"

	"github.com/marco-m/florist/pkg/florist"
)

// Input paths in the source fs.
const (
	InstallStaticFileSrc = "inst1.txt"
	InstallTplFileSrc    = "inst2.tmpl"
	ConfigTplFileSrc     = "config.txt.tpl"
)

// Output paths in the destination fs, relative to the customizable Flower.DstDir.
const (
	InstallStaticFileDst = "inst1.txt"
	InstallTmplFileDst   = "inst-tpl.txt"
	ConfigTplFileDst     = "config-tpl.txt"
)

type secrets struct {
	Fruit string
	//
	Secret string
	Custom string
}

var _ florist.Flower = (*Flower)(nil)

type Flower struct {
	// Base directory into which all files will be installed.
	DstDir string `default:"/tmp/daisy"`

	// kv, with default value, in the flower itself.
	Fruit string `default:"banana"`

	log  hclog.Logger
	user string
}

func (fl *Flower) String() string {
	return "daisy"
}

func (fl *Flower) Description() string {
	return "a daisy flower"
}

func (fl *Flower) Init() error {
	fl.log = florist.Log.ResetNamed(fl.String())

	if err := defaults.Set(fl); err != nil {
		return fmt.Errorf("%s.init: %s", fl, err)
	}

	var err error
	curUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}
	fl.user = curUser.Username

	return nil
}

func (fl *Flower) Install(files fs.FS, finder florist.Finder) error {
	log := fl.log.Named("install")
	log.Info("Creating dir", "dstDir", fl.DstDir, "user", fl.user)
	if err := florist.Mkdir(fl.DstDir, fl.user, 0755); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	dstPath1 := filepath.Join(fl.DstDir, InstallStaticFileDst)
	log.Debug("installing file (static)",
		"src", InstallStaticFileSrc, "dst", dstPath1)
	if err := florist.CopyFileFs(
		files, InstallStaticFileSrc, dstPath1, 0600, fl.user); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	// This shows how to use a customizable value (fl.Fruit) at install time.
	data := struct{ Fruit string }{Fruit: fl.Fruit}

	dstPath2 := filepath.Join(fl.DstDir, InstallTmplFileDst)
	log.Debug("installing file (templated)",
		"src", InstallTplFileSrc, "dst", dstPath2)
	if err := florist.CopyTemplateFs(
		files, InstallTplFileSrc, dstPath2, 0600, fl.user, data, "", ""); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	return nil
}

func (fl *Flower) Configure(files fs.FS, finder florist.Finder) error {
	log := fl.log.Named("configure")
	log.Debug("loading secrets")
	data := secrets{
		Fruit:  fl.Fruit,
		Secret: finder.Get("secret"),
		Custom: finder.Get("custom"),
	}
	if err := finder.Error(); err != nil {
		return fmt.Errorf("daisy.configure: %s", err)
	}

	dstPath := filepath.Join(fl.DstDir, ConfigTplFileDst)
	log.Debug("installing file (templated)", "src", ConfigTplFileSrc, "dst", dstPath)
	if err := florist.CopyTemplateFs(
		files, ConfigTplFileSrc, dstPath, 0600, fl.user, data, "", ""); err != nil {
		return fmt.Errorf("%s: %s", log.Name(), err)
	}

	return nil
}
