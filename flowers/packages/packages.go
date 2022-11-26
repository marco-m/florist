// Package packages contains a flower to add and remove vanilla packages.
package packages

import (
	"fmt"

	"github.com/hashicorp/go-hclog"
	"github.com/marco-m/florist"
	"github.com/marco-m/florist/pkg/apt"
)

type Flower struct {
	Name   string
	Add    []string
	Remove []string
	log    hclog.Logger
}

func (fl *Flower) String() string {
	return fl.Name
}

func (fl *Flower) Description() string {
	return "add and remove packages with the system package manager"
}

func (fl *Flower) Init() error {
	name := fmt.Sprintf("florist.flower.%s", fl)
	fl.log = florist.Log.ResetNamed(name)

	if fl.Name == "" {
		return fmt.Errorf("%s.init: missing name", name)
	}
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
