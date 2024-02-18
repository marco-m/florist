package apt_test

import (
	"testing"

	"github.com/marco-m/florist/pkg/apt"
	"github.com/marco-m/florist/pkg/florist"
)

func TestAptUpdateVM(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)

	err := apt.Update(florist.CacheValidity)

	if err != nil {
		t.Fatalf("\nhave: %s\nwant: <no error>", err)
	}
}

func TestAptInstallVM(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)

	err := apt.Install("ripgrep")

	if err != nil {
		t.Fatalf("\nhave: %s\nwant: <no error>", err)
	}
}
