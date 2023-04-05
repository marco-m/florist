// Program "florist" is the provisioner for the VM used to develop florist itself.
package main

import (
	"embed"
	"fmt"
	"io/fs"
	"os"

	"github.com/hashicorp/go-hclog"

	"github.com/marco-m/florist/flowers/consul"
	"github.com/marco-m/florist/flowers/consultemplate"
	"github.com/marco-m/florist/flowers/docker"
	"github.com/marco-m/florist/flowers/fishshell"
	"github.com/marco-m/florist/flowers/golang"
	"github.com/marco-m/florist/flowers/locale"
	"github.com/marco-m/florist/flowers/nomad"
	"github.com/marco-m/florist/flowers/ospackages"
	"github.com/marco-m/florist/flowers/taskfile"
	"github.com/marco-m/florist/pkg/florist"
	"github.com/marco-m/florist/pkg/provisioner"
)

//go:embed embed
var embedded embed.FS

func main() {
	log := florist.NewLogger("florist-prov")
	os.Exit(provisioner.Main(log, prepare))
}

func prepare(log hclog.Logger) (*provisioner.Provisioner, error) {
	files, err := fs.Sub(embedded, "embed/files")
	if err != nil {
		return nil, fmt.Errorf("prepare: %s", err)
	}
	secrets, err := fs.Sub(embedded, "embed/secrets")
	if err != nil {
		return nil, fmt.Errorf("prepare: %s", err)
	}

	// Create a provisioner.
	prov, err := provisioner.New(log, florist.CacheValidity, files, secrets)
	if err != nil {
		return nil, err
	}

	//
	// Add bouquets (bunches of flowers).
	//

	// FIXME I removed the copyfilesFlower, so fix comment or fix code!!!
	// the first flower (copyfilesFlower) belongs to more than one bouquet,
	// the second flower (consultemplate.Flower) is instantiated inline.
	if err := prov.AddBouquet("all-you-need", "install everything",
		&consultemplate.Flower{
			Version: "0.27.2",
			Hash:    "d3d428ede8cb6e486d74b74deb9a7cdba6a6de293f3311f178cc147f1d1837e8",
		}); err != nil {
		return nil, err
	}

	//
	// Some other bouquets, to show the available flowers in florist.
	//

	if err := prov.AddBouquet("nomadconsulclients", "install Nomad and Consul clients",
		&nomad.ClientFlower{
			Version: "1.4.2",
			Hash:    "6e24efd6dfba0ba2df31347753f615cae4d3632090e68fc90933e51e640f7afc",
		},
		&consul.ClientFlower{
			Version: "1.14.0",
			Hash:    "6907e0dc83a05acaa9de60217e44ce55bd05c98152dcef02f9258bd2a474f2b3",
		},
		&docker.Flower{},
	); err != nil {
		return nil, err
	}

	if err := prov.AddBouquet("dev", "install a development environment",
		&locale.Flower{
			Lang: locale.Lang_en_US_UTF8,
		},
		&ospackages.Flower{
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
		&taskfile.Flower{
			Version: "3.21.0",
			Hash:    "7232508b0040398b3dcce5d92dfe05f65723680eab2017f3cee6c0a7cf9dd6c1",
		},
		&golang.Flower{
			Version: "1.20.1",
			Hash:    "000a5b1fca4f75895f78befeb2eecf10bfff3c428597f3f1e69133b63b911b02",
		},
		&fishshell.Flower{
			Usernames:    []string{"vagrant"},
			SetAsDefault: true,
		},
	); err != nil {
		return nil, err
	}

	// Run the installer.
	return prov, nil
}
