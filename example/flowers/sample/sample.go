// Package sample is an example flower that copies, expanding the template, one
// file at install time and one file at configure time.
package sample

import (
	"fmt"
	"io/fs"
	"os/user"
	"path/filepath"

	"github.com/creasty/defaults"
	"github.com/hashicorp/go-hclog"

	"github.com/marco-m/florist"
)

// Input paths in the source fs.
const (
	InstallStaticFileSrc = "embed/files/sample/inst-static.txt"
	InstallTmplFileSrc   = "embed/files/sample/inst-tmpl.txt.tmpl"
	ConfigTmplFileSrc    = "embed/files/sample/config-tmpl.txt.tmpl"

	SecretK = "embed/secrets/secret"
	CustomK = "embed/secrets/custom"
)

// Output paths in the destination fs, relative to the customizable Flower.DstDir.
const (
	InstallStaticFileDst = "inst-static.txt"
	InstallTmplFileDst   = "inst-tmpl.txt"
	ConfigTmplFileDst    = "config-tmpl.txt"
)

var _ florist.Flower = (*Flower)(nil)

type Flower struct {
	// Used at install and at configure time: static and non-secret data.
	FilesFs fs.FS
	// Used only at configure time: dynamic or secret data.
	SecretsFs fs.FS

	// Base directory into which all files will be installed.
	DstDir string `default:"/opt/sample"`

	// kv, with default value, in the flower itself.
	Fruit string `default:"banana"`

	log  hclog.Logger
	user *user.User
}

func (fl *Flower) String() string {
	// When writing your own flower, replace "florist" with the name of your project.
	return "florist.sample"
}

func (fl *Flower) Description() string {
	return "a sample flower"
}

func (fl *Flower) Init() error {
	fl.log = florist.Log.ResetNamed(fl.String())

	if err := defaults.Set(fl); err != nil {
		return fmt.Errorf("%s.init: %s", fl, err)
	}

	var err error
	fl.user, err = user.Current()
	if err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	return nil
}

func (fl *Flower) Install() error {
	fl.log = fl.log.Named("install")

	fl.log.Info("Creating dir", "dstDir", fl.DstDir, "user", fl.user)
	if err := florist.Mkdir(fl.DstDir, fl.user, 0755); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	dstPath1 := filepath.Join(fl.DstDir, InstallStaticFileDst)
	fl.log.Debug("installing file (static)",
		"src", InstallStaticFileSrc, "dst", dstPath1)
	if err := florist.CopyFileFromFs(
		fl.FilesFs, InstallStaticFileSrc, dstPath1, 0600, fl.user); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	// This shows how to use a customizable value (fl.Fruit) at install time.
	data2 := map[string]string{"Fruit": fl.Fruit}
	dstPath2 := filepath.Join(fl.DstDir, InstallTmplFileDst)
	fl.log.Debug("installing file (templated)",
		"src", InstallTmplFileSrc, "dst", dstPath2)
	if err := florist.CopyFileTemplateFromFs(
		fl.FilesFs, InstallTmplFileSrc, dstPath2, 0600, fl.user, data2); err != nil {
		return fmt.Errorf("%s: %s", fl, err)
	}

	return nil
}

func (fl *Flower) Configure() error {
	fl.log = fl.log.Named("configure")

	fl.log.Debug("loading secrets")
	data3, err := florist.MakeTmplData(fl.SecretsFs, SecretK, CustomK)
	if err != nil {
		return fmt.Errorf("%s:\n%s", fl.log.Name(), err)
	}
	// FIXME, but how????
	data3["Fruit"] = fl.Fruit

	fmt.Println("data3", data3)

	dstPath3 := filepath.Join(fl.DstDir, ConfigTmplFileDst)
	fl.log.Debug("installing file (templated)",
		"src", ConfigTmplFileSrc, "dst", dstPath3)
	if err := florist.CopyFileTemplateFromFs(
		fl.FilesFs, ConfigTmplFileSrc, dstPath3, 0600, fl.user, data3); err != nil {
		return fmt.Errorf("%s: %s", fl.log.Name(), err)
	}

	return nil
}
