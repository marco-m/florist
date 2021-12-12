package main

import (
	"embed"
	"os"
	"time"

	"github.com/hashicorp/go-hclog"

	"github.com/marco-m/florist"
	"github.com/marco-m/florist/example/flowers/copyfiles"
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

	// Add flowers.
	if err := inst.AddFlower(&copyfiles.Flower{
		FilesFS:  filesFS,
		SrcFiles: []string{"hello.txt"},
	}); err != nil {
		return err
	}

	// Run the installer.
	return inst.Run()
}
