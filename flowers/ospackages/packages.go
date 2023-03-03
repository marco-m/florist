// Package ospackages contains a flower to add and remove packages with the OS package manager.
package ospackages

import (
	"fmt"
	"io/fs"

	"github.com/hashicorp/go-hclog"

	"github.com/marco-m/florist"
	"github.com/marco-m/florist/pkg/apt"
)

var _ florist.Flower = (*Flower)(nil)

type Flower struct {
	fsys   fs.FS
	Add    []string
	Remove []string
	log    hclog.Logger
}

func (fl *Flower) String() string {
	return "ospackages"
}

func (fl *Flower) Description() string {
	return "add/remove packages with the OS package manager"
}

func (fl *Flower) Init(fsys fs.FS) error {
	fl.fsys = fsys
	name := fmt.Sprintf("florist.flower.%s", fl)
	fl.log = florist.Log.ResetNamed(name)

	if len(fl.Add) == 0 && len(fl.Remove) == 0 {
		return fmt.Errorf("%s.init: missing packages", name)
	}

	return nil
}

func (fl *Flower) Install() error {
	fl.log.Info("begin")
	defer fl.log.Info("end")

	if len(fl.Add) > 0 {
		fl.log.Info("adding packages")
		if err := apt.Install(fl.Add...); err != nil {
			return err
		}
	}

	if len(fl.Remove) > 0 {
		fl.log.Info("removing packages")
		if err := apt.Remove(fl.Remove...); err != nil {
			return err
		}
	}

	return nil
}

func (fl *Flower) Configure() error {
	return nil
}
