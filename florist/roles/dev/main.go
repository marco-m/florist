// This program is the provisioner for the VM used to develop florist itself.
package main

import (
	"os"

	"github.com/marco-m/florist/flowers/fishshell"
	"github.com/marco-m/florist/flowers/golang"
	"github.com/marco-m/florist/flowers/locale"
	"github.com/marco-m/florist/flowers/ospackages"
	"github.com/marco-m/florist/flowers/task"
	"github.com/marco-m/florist/pkg/florist"
)

func main() {
	os.Exit(florist.MainInt(&florist.Options{
		SetupFn:        setup,
		PreConfigureFn: preConfigure,
	}))
}

func setup(prov *florist.Provisioner) error {
	return prov.AddFlowers(
		&locale.Flower{
			Inst: locale.Inst{
				Lang: locale.Lang_en_US_UTF8,
			},
		},
		&ospackages.Flower{
			Inst: ospackages.Inst{
				Add: []string{
					"build-essential",
					"sntp",
					"ripgrep",
					"rsync", // Needed by Jetbrains Goland SSH run target.
				},
				Remove: []string{
					"unattended-upgrades",
				},
			},
		},
		&task.Flower{
			Inst: task.Inst{
				Version: "3.21.0",
				Hash:    "7232508b0040398b3dcce5d92dfe05f65723680eab2017f3cee6c0a7cf9dd6c1",
			},
		},
		&golang.Flower{
			Inst: golang.Inst{
				Version: "1.22.0",
				Hash:    "f6c8a87aa03b92c4b0bf3d558e28ea03006eb29db78917daec5cfb6ec1046265",
			},
		},
		&fishshell.Flower{
			Inst: fishshell.Inst{
				Usernames:    []string{"vagrant"},
				SetAsDefault: true,
			},
		},
	)
}

func preConfigure(prov *florist.Provisioner, config *florist.Config) (any, error) {
	return nil, nil
}
