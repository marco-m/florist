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
				Version: "3.44.0",
				Hash:    "d6c9c0a14793659766ee0c06f9843452942ae6982a3151c6bbd78959c1682b82",
			},
		},
		&golang.Flower{
			Inst: golang.Inst{
				Version: "1.24.4",
				Hash:    "77e5da33bb72aeaef1ba4418b6fe511bc4d041873cbf82e5aa6318740df98717",
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
