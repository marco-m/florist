// This program is a small florist provisioner.
package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/marco-m/florist/example/flowers/daisy"
	"github.com/marco-m/florist/example/flowers/mint"
	"github.com/marco-m/florist/pkg/florist"
	"github.com/marco-m/florist/pkg/provisioner"
)

func main() {
	if err := mainErr(); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}

func mainErr() error {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	if err := florist.Init(log); err != nil {
		return err
	}
	return provisioner.Main(setup, configure)
}

func setup() (*provisioner.Provisioner, error) {
	prov, err := provisioner.New(florist.CacheValidity)
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
