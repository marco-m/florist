package apt_test

import (
	"testing"

	"github.com/marco-m/florist"
	"github.com/marco-m/florist/pkg/apt"
)

func TestAptUpdateVM(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)

	err := apt.Update(florist.CacheValidityDefault)

	if err != nil {
		t.Fatalf("\nhave: %s\nwant: <no error>", err)
	}
}

func TestAptInstallVM(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)

	err := apt.Install("netcat")

	if err != nil {
		t.Fatalf("\nhave: %s\nwant: <no error>", err)
	}
}
