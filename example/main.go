package main

import (
	"embed"
	"os"
	"time"

	"github.com/hashicorp/go-hclog"

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
	log := florist.NewLogger()
	err := run(log)
	elapsed := time.Since(start).Round(time.Millisecond)

	if err != nil {
		log.Error("exit failure", "error", err, "elapsed", elapsed)
		os.Exit(1)
	}
	log.Info("exit success", "elapsed", elapsed)
}

func run(log hclog.Logger) error {
	// Create an installer.
	inst := installer.New(log, florist.CacheValidityDefault)

	//
	// Add bouquets (bunches of flowers).
	//

	// The simplest possible: 1 flower.
	//
	bouquet1 := []florist.Flower{
		&copyfiles.Flower{
			FilesFS:  filesFS,
			SrcFiles: []string{"hello.txt"},
		},
	}

	// To keep the original name and description of the flower, just pass empty
	// strings for the first two parameters (this makes sense only when there
	// is a single flower in the bouquet).
	// Running `example-florist list` will print:
	// copyfiles            copy files from an embed.FS to the real filesystem
	if err := inst.AddBouquet("", "", bouquet1); err != nil {
		return err
	}

	// To customize the name and description, set the first two parameters.
	// Running `example-florist list` will print:
	// files                install the files for projectX
	if err := inst.AddBouquet("files", "install the files for projectX",
		bouquet1); err != nil {
		return err
	}

	// A bouquet that bundles other bouquets
	//
	bouquet2 := []florist.Flower{&consultemplate.Flower{
		FilesFS: filesFS,
		Version: "0.27.2",
		Hash:    "d3d428ede8cb6e486d74b74deb9a7cdba6a6de293f3311f178cc147f1d1837e8",
	}}
	bouquet2 = append(bouquet2, bouquet1...)
	if err := inst.AddBouquet("all-you-need", "install everything",
		bouquet2); err != nil {
		return err
	}

	//
	// Some other bouquets, to show the available flowers in florist.
	//

	if err := inst.AddBouquet("nomadconsulclients", "install Nomad and Consul clients",
		[]florist.Flower{
			&nomad.ClientFlower{},
			&consul.ClientFlower{},
			&docker.Flower{},
		}); err != nil {
		return err
	}

	if err := inst.AddBouquet("nomadconsulservers", "install Nomad and Consul servers",
		[]florist.Flower{
			&nomad.ServerFlower{},
			&consul.ServerFlower{},
		}); err != nil {
		return err
	}

	// Run the installer.
	return inst.Run()
}
