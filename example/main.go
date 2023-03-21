// Program "example" is a small florist installer.
// For a bigger example, see program "installer", still in the florist module.
package main

import (
	"embed"
	"fmt"
	"github.com/marco-m/florist/example/flowers/daisy"
	"io/fs"
	"os"

	"github.com/hashicorp/go-hclog"

	"github.com/marco-m/florist"
	"github.com/marco-m/florist/pkg/installer"
)

//go:embed embed
var embedded embed.FS

func main() {
	log := florist.NewLogger("example")
	os.Exit(installer.Main(log, prepare))
}

func prepare(log hclog.Logger) (*installer.Installer, error) {
	files, err := fs.Sub(embedded, "embed/files")
	if err != nil {
		return nil, fmt.Errorf("prepare: %s", err)
	}
	secrets, err := fs.Sub(embedded, "embed/secrets")
	if err != nil {
		return nil, fmt.Errorf("prepare: %s", err)
	}

	// Create an installer.
	inst, err := installer.New(log, florist.CacheValidity, files, secrets)
	if err != nil {
		return nil, err
	}
	inst.UseWorkdir()

	// Handle a flower without secrets
	{
		// Create the daisy flower.
		flower := &daisy.Flower{Fruit: "strawberry"}

		// Add a bouquet.
		if err := inst.AddBouquet("mango", "a mango bouquet", flower); err != nil {
			return nil, err
		}
	}

	// Handle a flower with secrets
	{
		// FIXME WRITEME
	}

	// FIXME SO: Must clarify in my mind what is an instance and what is an image for a specific
	// role (nothing to do with Ansible roles here!), how many bouquest can go on the
	// same instance, wether it is optional or required that the executbale is build
	// specifically for each instance and so on.

	// Return the installer filled with bouquets.
	return inst, nil
}
