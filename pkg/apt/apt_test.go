package apt_test

import (
	"testing"

	"github.com/marco-m/florist"
	"github.com/marco-m/florist/pkg/apt"
	"github.com/marco-m/xprog"
)

func TestAptUpdateVM(t *testing.T) {
	if xprog.Absent() {
		t.Skip("skip: test requires xprog")
	}

	err := apt.Update(florist.CacheValidityDefault)

	if err != nil {
		t.Fatalf("\nhave: %s\nwant: <no error>", err)
	}
}

func TestAptInstallVM(t *testing.T) {
	if xprog.Absent() {
		t.Skip("skip: test requires xprog")
	}

	err := apt.Install("netcat")

	if err != nil {
		t.Fatalf("\nhave: %s\nwant: <no error>", err)
	}
}
