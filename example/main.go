// Program "one" is a small florist provisioner.
package main

import (
	"embed"
	"fmt"
	"io/fs"
	"os"

	"github.com/hashicorp/go-hclog"

	"github.com/marco-m/florist/example/flowers/daisy"
	"github.com/marco-m/florist/pkg/florist"
	"github.com/marco-m/florist/pkg/provisioner"
)

//go:embed embed
var embedded embed.FS

func main() {
	log := florist.NewLogger("example")
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
	prov.UseWorkdir()

	// Handle a flower without secrets
	{
		// Create the daisy flower.
		flower := &daisy.Flower{Fruit: "strawberry"}

		// Add a bouquet.
		if err := prov.AddBouquet("mango", "a mango bouquet", flower); err != nil {
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
	return prov, nil
}
