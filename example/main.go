// This program is a small 🌼 florist 🌺 provisioner.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"example/flowers/daisy"
	"example/flowers/mint"

	"github.com/marco-m/florist/pkg/florist"
)

func main() {
	os.Exit(florist.MainInt(&florist.Options{
		SetupFn:        setup,
		PreConfigureFn: preConfigure,
		//
		// Sometimes the OS default tempDir is too small. You can use TempDir
		// to override.
		// TempDir: "/path/with/enough/space",
		//
		// WARNING: RootDir is to be set ONLY for test. Here we set it
		// ONLY BECAUSE this is a sample provisioner!!!
		RootDir: filepath.Join(os.TempDir(), "example-florist"),
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
