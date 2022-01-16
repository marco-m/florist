// Package test is a test flower
package test

import (
	"fmt"
	"os"
	"os/user"
	"path"

	"github.com/marco-m/florist"
)

type Flower struct {
	Contents string
	Dst      string
}

func (fl *Flower) String() string {
	return "test"
}

func (fl *Flower) Description() string {
	return "a test flower"
}

func (fl *Flower) Install() error {
	log := florist.Log.ResetNamed("florist.flower.test")
	log.Info("begin")
	defer log.Info("end")

	curUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("%s: %s", log.Name(), err)
	}

	dstDir := path.Dir(fl.Dst)
	log.Info("Create dir", "dstDir", dstDir)
	if err := florist.Mkdir(dstDir, curUser, 0755); err != nil {
		return fmt.Errorf("%s: %s", log.Name(), err)
	}

	log.Debug("writing file", "dst", fl.Dst, "contents", fl.Contents)
	if err := os.WriteFile(fl.Dst, []byte(fl.Contents), 0644); err != nil {
		return fmt.Errorf("%s: %s", log.Name(), err)
	}

	return nil
}
