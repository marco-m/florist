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
		SetupFn:        setup,
		PreConfigureFn: preConfigure,
	}))
}

func setup(prov *florist.Provisioner) error {
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

func preConfigure(prov *florist.Provisioner, config *florist.Config) (any, error) {
	prov.Flowers()[daisy.Name].(*daisy.Flower).Conf = daisy.Conf{
		Environment: config.Get("Environment"),
		GossipKey:   config.Get("GossipKey"),
	}

	prov.Flowers()[mint.Name].(*mint.Flower).Conf = mint.Conf{
		Aroma: config.Get("Aroma"), // dynamic setting
	}

	return nil, florist.JoinErrors(config.Errors())
}
