package provisioner_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/rogpeppe/go-internal/testscript"

	"github.com/marco-m/florist/pkg/florist"
	"github.com/marco-m/florist/pkg/provisioner"
)

func setup1(log hclog.Logger) (*provisioner.Provisioner, error) {
	prov, err := provisioner.New(log, florist.CacheValidity)
	if err != nil {
		return nil, fmt.Errorf("setup: %s", err)
	}
	prov.UseWorkdir() // FIXME what is this???

	if err := prov.AddFlowers(
		&TestFlower{Inst: Inst{name: "A", desc: "desc A"}},
		&TestFlower{Inst: Inst{name: "B", desc: "desc B"}},
	); err != nil {
		return nil, fmt.Errorf("setup: %s", err)
	}

	return prov, nil
}

func configure1(prov *provisioner.Provisioner, config *florist.Config) error {
	return nil
}

func TestScriptProvisioner(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "testdata",
	})
}

func TestMain(m *testing.M) {
	log := florist.NewLogger("test")
	os.Exit(testscript.RunMain(m, map[string]func() int{
		"provisioner1": func() int { return provisioner.Main(log, setup1, configure1) },
		//"provisioner2": func() int { return provisioner.Main(log, setup2) },
	}))
}
