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
			Lang: locale.Lang_en_US_UTF8,
		},
		&ospackages.Flower{
			Add: []string{
				//"build-essential",
				// "sntp",
				"ripgrep",
				"rsync", // Needed by Jetbrains Goland SSH run target.
			},
			Remove: []string{
				"unattended-upgrades",
			},
		},
		&task.Flower{
			Version: "3.54.0",
			Hash:    "680859dbb4d881a9c72d4d9a8f510825450849af8567deacd7302c01124416fb",
		},
		&golang.Flower{
			Version: "1.27.1",
			Hash:    "63d339f0da5ab53635a56f2490a7984dfe12dfcff22ad749f63edaf590168445",
		},
		&fishshell.Flower{
			Usernames: []string{"root", "vagrant"},
			// For some reasons, Fish as default shell breaks some non-interactive
			// usages with ssh.
			SetAsDefault: false,
		},
	)
}

func preConfigure(prov *florist.Provisioner, config *florist.Config) (any, error) {
	return nil, nil
}
