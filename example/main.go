// Program "example" is the smallest possible florist installer.
// For a bigger example, see program "installer", still in the florist module.
package main

import (
	"embed"
	"os"
	"time"

	"github.com/hashicorp/go-hclog"

	"github.com/marco-m/florist"
	"github.com/marco-m/florist/example/flowers/sample"
	"github.com/marco-m/florist/pkg/installer"
)

//go:embed embed/files
var filesFS embed.FS

//go:embed embed/secrets
var secretsFS embed.FS

func main() {
	start := time.Now()
	log := florist.NewLogger("example")
	err := run(log)
	elapsed := time.Since(start).Round(time.Millisecond)

	if err != nil {
		log.Error("", "exit", "failure", "error", err, "elapsed", elapsed)
		os.Exit(1)
	}
	log.Info("", "exit", "success", "elapsed", elapsed)
}

func run(log hclog.Logger) error {
	// Create an installer.
	inst := installer.New(log, florist.CacheValidityDefault, filesFS, secretsFS)

	// Create the sample flower.
	flower := sample.Flower{
		FilesFs:   filesFS,
		SecretsFs: secretsFS,
		Fruit:     "strawberry",
	}

	// Add a bouquet with a single flower.
	if err := inst.AddFlower(&flower); err != nil {
		return err
	}

	// Run the installer.
	return inst.Run()
}
