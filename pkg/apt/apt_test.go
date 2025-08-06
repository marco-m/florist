package apt_test

import (
	"io"
	"testing"
	"time"

	"github.com/marco-m/florist/pkg/apt"
	"github.com/marco-m/florist/pkg/florist"
	"github.com/marco-m/florist/pkg/provisioner"
	"github.com/marco-m/rosina/assert"
)

func TestAptUpdateVM(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)

	err := provisioner.LowLevelInit(io.Discard, "INFO", time.Hour)
	assert.NoError(t, err, "provisioner.LowLevelInit")

	err = apt.Update(provisioner.CacheValidity())
	assert.NoError(t, err, "apt.Update")
}

func TestAptInstallVM(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)

	err := apt.Install("ripgrep")
	assert.NoError(t, err, "apt.Install")
}
