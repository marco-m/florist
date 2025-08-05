// Package timezone contains a flower to setup the timezone
package timezone

import (
	"fmt"
	"log/slog"
	"os"
	"path"

	"github.com/creasty/defaults"

	"github.com/marco-m/florist/pkg/florist"
)

const zoneinfoDir = "/usr/share/zoneinfo"

const Name = "timezone"

var _ florist.Flower = (*Flower)(nil)

type Flower struct {
	Inst
	Conf
}

type Inst struct {
	// The timezone to set, in the same format as the TZ environment variable.
	// For example: "Europe/Zurich".
	// To list the valid timezones, see directory /usr/share/zoneinfo on a Unix
	// system.
	Timezone string
}

type Conf struct{}

func (fl *Flower) String() string {
	return Name
}

func (fl *Flower) Description() string {
	return "setup the timezone as localtime (leave the RTC on UTC)"
}

func (fl *Flower) Embedded() []string {
	return nil
}

func (fl *Flower) Init() error {
	if err := defaults.Set(fl); err != nil {
		return fmt.Errorf("%s: %s", Name, err)
	}
	if fl.Timezone == "" {
		return fmt.Errorf("%s.init: missing timezone", Name)
	}
	// Detect dot-dot and other path syntax that should not be there!
	cleaned := path.Clean(fl.Timezone)
	if cleaned != fl.Timezone {
		return fmt.Errorf("%s.init: timezone: %q; want: %q", Name, fl.Timezone, cleaned)
	}
	return nil
}

func (fl *Flower) Install() error {
	log := slog.With("flower", Name+".install")

	// We leave the RTC on UTC and modify only the local timezone.
	// See "man hwclock" on Linux for a detailed explanation.

	// Names are the same as "man ln": ln -s target linkname
	target := path.Join(zoneinfoDir, fl.Timezone)
	linkname := "/etc/localtime"

	// Before removing the current localtime to allow the symlink, let's
	// double-check that the wanted timezone exists. As usual with file
	// existence checks, this is racy, but benign in this case.

	targetExists, err := florist.FileExists(target)
	if err != nil {
		return fmt.Errorf("%s.install: checking if %q exists: %s", Name, target, err)
	}
	if !targetExists {
		return fmt.Errorf("%s.install: timezone %q does not exist: %s", Name, target, err)
	}

	// Remove the current linkname to allow the symlink. This might fail for
	// multiple reasons. We hope for the best.
	err = os.Remove(linkname)
	if err != nil {
		log.Debug("remove-linkname", "linkname", linkname, "err", err)
	}

	if err := os.Symlink(target, linkname); err != nil {
		return fmt.Errorf("%s.install: symlink: %s", Name, err)
	}
	log.Info("installed", "timezone-localtime", fl.Timezone)

	return nil
}

func (fl *Flower) Configure() error {
	log := slog.With("flower", Name+".configure")
	log.Debug("nothing to do")
	return nil
}
