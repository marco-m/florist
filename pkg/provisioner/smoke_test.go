package provisioner_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"

	"github.com/marco-m/florist/pkg/florist"
	"github.com/marco-m/florist/pkg/provisioner"
)

func setup1() (*provisioner.Provisioner, error) {
	prov, err := provisioner.New(florist.CacheValidity)
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
	os.Exit(testscript.RunMain(m, map[string]func() int{
		"provisioner1": func() int {
			if err := provisioner.Main(setup1, configure1); err != nil {
				fmt.Println("error:", err)
				return 1
			}
			return 0
		},
		//"provisioner2": func() int { return provisioner.Main(log, setup2) },
	}))
}
