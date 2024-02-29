package apt_test

import (
	"io"
	"testing"
	"time"

	"github.com/go-quicktest/qt"

	"github.com/marco-m/florist/pkg/apt"
	"github.com/marco-m/florist/pkg/florist"
)

func TestAptUpdateVM(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)

	err := florist.LowLevelInit(io.Discard, "INFO", time.Hour)
	qt.Assert(t, qt.IsNil(err))

	err = apt.Update(florist.CacheValidity())
	qt.Assert(t, qt.IsNil(err))
}

func TestAptInstallVM(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)

	err := apt.Install("ripgrep")
	qt.Assert(t, qt.IsNil(err))
}
