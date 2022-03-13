// Package packages contains a flower to install vanilla packages.
package packages

import (
	"fmt"

	"github.com/hashicorp/go-hclog"
	"github.com/marco-m/florist"
	"github.com/marco-m/florist/pkg/apt"
)

type Options struct {
	Name     string
	Packages []string
}

type Flower struct {
	Options
	log hclog.Logger
}

func New(opts Options) (*Flower, error) {
	fl := Flower{Options: opts}
	name := fmt.Sprintf("florist.flower.%s", fl)
	fl.log = florist.Log.ResetNamed(name)

	if fl.Name == "" {
		return nil, fmt.Errorf("%s.new: missing name", name)
	}
	if len(fl.Packages) == 0 {
		return nil, fmt.Errorf("%s.new: missing packages", name)
	}

	return &fl, nil
}

func (fl Flower) String() string {
	return fl.Name
}

func (fl Flower) Description() string {
	return "install packages with the system package manager"
}

func (fl *Flower) Install() error {
	fl.log.Info("begin")
	defer fl.log.Info("end")

	fl.log.Info("Install packages")
	if err := apt.Install(fl.Packages...); err != nil {
		return err
	}

	return nil
}
