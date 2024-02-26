// Package ospackages contains a flower to add and remove packages with the OS package manager.
package ospackages

import (
	"fmt"

	"github.com/creasty/defaults"

	"github.com/marco-m/florist/pkg/apt"
	"github.com/marco-m/florist/pkg/florist"
)

const Name = "ospackages"

var _ florist.Flower = (*Flower)(nil)

type Flower struct {
	Inst
	Conf
}

type Inst struct {
	Add    []string
	Remove []string
}

type Conf struct{}

func (fl *Flower) String() string {
	return Name
}

func (fl *Flower) Description() string {
	return "add/remove packages with the OS package manager"
}

func (fl *Flower) Embedded() []string {
	return nil
}

func (fl *Flower) Init() error {
	if err := defaults.Set(fl); err != nil {
		return fmt.Errorf("%s: %s", Name, err)
	}

	if len(fl.Add) == 0 && len(fl.Remove) == 0 {
		return fmt.Errorf("%s.init: missing packages", Name)
	}

	return nil
}

func (fl *Flower) Install() error {
	log := florist.Log.With("flower", Name+".install")

	if len(fl.Add) > 0 {
		log.Info("adding packages")
		if err := apt.Install(fl.Add...); err != nil {
			return err
		}
	}

	if len(fl.Remove) > 0 {
		log.Info("removing packages")
		if err := apt.Remove(fl.Remove...); err != nil {
			return err
		}
	}

	return nil
}

func (fl *Flower) Configure() error {
	log := florist.Log.With("flower", Name+".configure")
	log.Debug("nothing to do")
	return nil
}
