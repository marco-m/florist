// This program is a small florist provisioner.
package main

import (
	"fmt"
	"os"

	"github.com/hashicorp/go-hclog"

	"github.com/marco-m/florist/example/flowers/daisy"
	"github.com/marco-m/florist/example/flowers/mint"
	"github.com/marco-m/florist/pkg/florist"
	"github.com/marco-m/florist/pkg/provisioner"
)

func main() {
	log := florist.NewLogger("example")
	os.Exit(provisioner.Main(log, setup, configure))
}

func setup(log hclog.Logger) (*provisioner.Provisioner, error) {
	prov, err := provisioner.New(log, florist.CacheValidity)
	if err != nil {
		return nil, fmt.Errorf("setup: %s", err)
	}
	prov.UseWorkdir() // FIXME what is this???

	err = prov.AddFlowers(
		&daisy.Flower{
			Inst: daisy.Inst{
				PetalColor: "blue", // embedded setting
			},
		},
		&mint.Flower{},
	)
	if err != nil {
		return nil, fmt.Errorf("setup: %s", err)
	}

	return prov, nil
}

func configure(prov *provisioner.Provisioner, config *florist.Config) error {
	prov.Flowers()[daisy.Name].(*daisy.Flower).Conf = daisy.Conf{
		Environment: config.Get("Environment"),
		GossipKey:   config.Get("GossipKey"),
	}

	prov.Flowers()[mint.Name].(*mint.Flower).Conf = mint.Conf{
		Aroma: config.Get("Aroma"), // dynamic setting
	}

	return nil
}
