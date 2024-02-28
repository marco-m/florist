// This program is a small florist provisioner.
package main

import (
	"fmt"
	"os"

	"github.com/marco-m/florist/example/flowers/daisy"
	"github.com/marco-m/florist/example/flowers/mint"
	"github.com/marco-m/florist/pkg/florist"
)

func main() {
	os.Exit(florist.MainInt(&florist.Options{
		SetupFn:     setup,
		ConfigureFn: configure,
	}))
}

func setup(prov *florist.Provisioner) error {
	prov.UseWorkdir() // FIXME what is this???
	err := prov.AddFlowers(
		&daisy.Flower{
			Inst: daisy.Inst{
				PetalColor: "blue", // embedded setting
			},
		},
		&mint.Flower{},
	)
	if err != nil {
		return fmt.Errorf("setup: %s", err)
	}
	return nil
}

func configure(prov *florist.Provisioner, config *florist.Config) error {
	prov.Flowers()[daisy.Name].(*daisy.Flower).Conf = daisy.Conf{
		Environment: config.Get("Environment"),
		GossipKey:   config.Get("GossipKey"),
	}

	prov.Flowers()[mint.Name].(*mint.Flower).Conf = mint.Conf{
		Aroma: config.Get("Aroma"), // dynamic setting
	}

	return florist.JoinErrors(config.Errors())
}
