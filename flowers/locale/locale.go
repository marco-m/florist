// Package locale contains a flower to setup the locale
package locale

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/creasty/defaults"

	"github.com/marco-m/florist/pkg/apt"
	"github.com/marco-m/florist/pkg/florist"
)

const (
	Lang_en_US_UTF8 = "en_US.UTF-8"
)

const Name = "locale"

var _ florist.Flower = (*Flower)(nil)

type Flower struct {
	Inst
	Conf
}

type Inst struct {
	Lang string // the LANG of the locale.
}

type Conf struct {
}

func (fl *Flower) String() string {
	return Name
}

func (fl *Flower) Description() string {
	return "setup locale"
}

func (fl *Flower) Embedded() []string {
	return nil
}

func (fl *Flower) Init() error {
	if err := defaults.Set(fl); err != nil {
		return fmt.Errorf("%s: %s", Name, err)
	}
	if fl.Lang == "" {
		return fmt.Errorf("%s.init: missing lang", Name)
	}
	return nil
}

func (fl *Flower) Install() error {
	log := florist.Log.ResetNamed(Name + ".install")

	log.Info("Install needed packages")
	if err := apt.Install("locales"); err != nil {
		return err
	}

	log.Info("Setup locale", "lang", fl.Lang)
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
		log.Info("locale already present, skipping generation", "lang", fl.Lang)
		return nil
	}

	locale := fmt.Sprintf("%s UTF-8\n", fl.Lang)
	if err := os.WriteFile("/etc/locale.gen", []byte(locale), 0644); err != nil {
		return err
	}
	cmd = exec.Command("locale-gen")
	if err := florist.CmdRun(log, cmd); err != nil {
		return err
	}
	return nil
}

func (fl *Flower) Configure() error {
	log := florist.Log.ResetNamed(Name + ".configure")
	log.Debug("nothing to do")
	return nil
}
