// Package locale contains a flower to setup the locale
package locale

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/hashicorp/go-hclog"
	"github.com/marco-m/florist"
	"github.com/marco-m/florist/pkg/apt"
)

const (
	Lang_en_US_UTF8 = "en_US.UTF-8"
)

type Flower struct {
	Lang string // the LANG of the locale.
	log  hclog.Logger
}

func (fl *Flower) String() string {
	return "locale"
}

func (fl *Flower) Description() string {
	return "setup locale"
}

func (fl *Flower) Init() error {
	name := fmt.Sprintf("florist.flower.%s", fl)
	fl.log = florist.Log.ResetNamed(name)

	if fl.Lang == "" {
		return fmt.Errorf("%s.new: missing lang", name)
	}

	return nil
}

func (fl *Flower) Install() error {
	fl.log.Info("begin")
	defer fl.log.Info("end")

	fl.log.Info("Install needed packages")
	if err := apt.Install("locales"); err != nil {
		return err
	}

	fl.log.Info("Setup locale", "flower", fl)
	// Since running locale-gen takes seconds, avoid if possible.
	cmd := exec.Command("localedef", "--list-archive")
	out, err := cmd.Output()
	if err != nil {
		return err
	}
	// For some unfathomable reasons, the output of "localedef --list-archive"
	// is "en_US.utf8", while LANG is "en_US.UTF-8" :-/
	left := strings.Split(fl.Lang, ".")[0]
	if strings.Contains(string(out), left) {
		fl.log.Info("locale already present, skipping generation", "lang", fl.Lang)
	} else {
		locale := fmt.Sprintf("%s UTF-8\n", fl.Lang)
		if err := os.WriteFile("/etc/locale.gen", []byte(locale), 0644); err != nil {
			return err
		}
		cmd = exec.Command("locale-gen")
		if err := florist.CmdRun(fl.log, cmd); err != nil {
			return err
		}
	}

	return nil

}

func (fl *Flower) Configure() error {
	return nil
}
