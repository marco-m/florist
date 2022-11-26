// Package test is a test flower
package test

import (
	"fmt"
	"os"
	"os/user"
	"path"

	"github.com/hashicorp/go-hclog"
	"github.com/marco-m/florist"
)

type Flower struct {
	Contents string
	DstPath  string
	log      hclog.Logger
}

func (fl *Flower) String() string {
	return "test"
}

func (fl *Flower) Description() string {
	return "a test flower"
}

func (fl *Flower) Init() error {
	fl.log = florist.Log.ResetNamed("florist.flower.test")
	return nil
}

func (fl *Flower) Install() error {
	fl.log.Info("begin")
	defer fl.log.Info("end")

	curUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("%s: %s", fl.log.Name(), err)
	}

	dstDir := path.Dir(fl.DstPath)
	fl.log.Info("Create dir", "dstDir", dstDir)
	if err := florist.Mkdir(dstDir, curUser, 0755); err != nil {
		return fmt.Errorf("%s: %s", fl.log.Name(), err)
	}

	fl.log.Debug("writing file", "dst", fl.DstPath, "contents", fl.Contents)
	if err := os.WriteFile(fl.DstPath, []byte(fl.Contents), 0644); err != nil {
		return fmt.Errorf("%s: %s", fl.log.Name(), err)
	}

	return nil
}
