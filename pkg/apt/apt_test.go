package apt_test

import (
	"io"
	"testing"
	"time"

	"github.com/marco-m/florist/pkg/apt"
	"github.com/marco-m/florist/pkg/florist"
	"github.com/marco-m/rosina"
)

func TestAptUpdateVM(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)

	err := florist.LowLevelInit(io.Discard, "INFO", time.Hour)
	rosina.AssertNoError(t, err)

	err = apt.Update(florist.CacheValidity())
	rosina.AssertNoError(t, err)
}

func TestAptInstallVM(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)

	err := apt.Install("ripgrep")
	rosina.AssertNoError(t, err)
}
