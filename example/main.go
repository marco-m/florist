// Program example shows how to write a florist installer.
package main

import (
	"embed"
	"io/fs"
	"os"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/marco-m/florist/flowers/fishshell"
	"github.com/marco-m/florist/flowers/golang"
	"github.com/marco-m/florist/flowers/locale"
	"github.com/marco-m/florist/flowers/packages"
	"github.com/marco-m/florist/flowers/taskfile"

	"github.com/marco-m/florist"
	"github.com/marco-m/florist/example/flowers/copyfiles"
	"github.com/marco-m/florist/flowers/consul"
	"github.com/marco-m/florist/flowers/consultemplate"
	"github.com/marco-m/florist/flowers/docker"
	"github.com/marco-m/florist/flowers/nomad"
	"github.com/marco-m/florist/pkg/installer"
)

//go:embed files
var filesFS embed.FS

func main() {
	start := time.Now()
	log := florist.NewLogger("florist-example")
	err := run(log)
	elapsed := time.Since(start).Round(time.Millisecond)

	if err != nil {
		log.Error("exit failure", "error", err, "elapsed", elapsed)
		os.Exit(1)
	}
	log.Info("exit success", "elapsed", elapsed)
}

func run(log hclog.Logger) error {
	filesFS, err := fs.Sub(filesFS, "files")
	if err != nil {
		return err
	}

	// Create an installer.
	inst := installer.New(log, florist.CacheValidityDefault, filesFS)

	//
	// Add bouquets (bunches of flowers).
	//

	// We instantiate this flower (instead of adding it directly to a bouquet as the
	// other flowers in this example) because we want to add it to multiple bouquets.
	copyfilesFlower := &copyfiles.Flower{
		FilesFS:  filesFS,
		SrcFiles: []string{"hello.txt"},
	}

	// A bouquet with a single flower (copyfilesFlower).
	// Although it contains a single flower, since we want to specify a custom name and
	// description, we use AddBouquet.
	// Running `example list` will print:
	// files                install the files for projectX
	if err := inst.AddBouquet("files", "install the files for projectX",
		copyfilesFlower); err != nil {
		return err
	}

	// Another bouquet with a single flower (copyfilesFlower).
	// In this case we accept the default name and description of the flower itself, so
	// we use the simpler AddFlower.
	// Running `example list` will print:
	// FIXME
	if err := inst.AddFlower(copyfilesFlower); err != nil {
		return err
	}

	// A bouquet with multiple flowers:
	// the first flower (copyfilesFlower) belongs to more than one bouquet,
	// the second flower (consultemplate.Flower) is instantiated inline.
	if err := inst.AddBouquet("all-you-need", "install everything",
		copyfilesFlower,
		&consultemplate.Flower{
			FilesFS: filesFS,
			Version: "0.27.2",
			Hash:    "d3d428ede8cb6e486d74b74deb9a7cdba6a6de293f3311f178cc147f1d1837e8",
		}); err != nil {
		return err
	}

	//
	// Some other bouquets, to show the available flowers in florist.
	//

	if err := inst.AddBouquet("nomadconsulclients", "install Nomad and Consul clients",
		&nomad.ClientFlower{
			FilesFS: filesFS,
			Version: "1.4.2",
			Hash:    "6e24efd6dfba0ba2df31347753f615cae4d3632090e68fc90933e51e640f7afc",
		},
		&consul.ClientFlower{
			FilesFS: filesFS,
			Version: "1.14.0",
			Hash:    "6907e0dc83a05acaa9de60217e44ce55bd05c98152dcef02f9258bd2a474f2b3",
		},
		&docker.Flower{},
	); err != nil {
		return err
	}

	if err := inst.AddBouquet("dev", "install a development environment",
		&locale.Flower{
			Lang: locale.Lang_en_US_UTF8,
		},
		&packages.Flower{
			Name: "dev",
			Add: []string{
				"build-essential",
				"sntp",
			},
			Remove: []string{
				"unattended-upgrades",
			},
		},
		&taskfile.Flower{
			Version: "3.18.0",
			Hash:    "b8bb5258d5fa3f0e278309b393b67a56065c0fa0e69be73e110b45094fa1e01c",
		},
		&golang.Flower{
			Version: "1.19.2",
			Hash:    "5e8c5a74fe6470dd7e055a461acda8bb4050ead8c2df70f227e3ff7d8eb7eeb6",
		},
		&fishshell.Flower{
			FilesFS:   filesFS,
			Usernames: []string{"vagrant"},
		},
	); err != nil {
		return err
	}

	// Run the installer.
	return inst.Run()
}
