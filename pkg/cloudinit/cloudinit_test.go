package cloudinit_test

import (
	"encoding/json"
	"testing"

	"github.com/marco-m/florist/pkg/cloudinit"
	"github.com/marco-m/rosina"
)

func TestFoo(t *testing.T) {
	type X struct {
		Ciccio string
		Bello  string
	}
	floristSettings := X{
		Ciccio: "aspide",
		Bello:  "zuppa",
	}

	floristConfigJson, err := json.Marshal(floristSettings)
	rosina.AssertNoError(t, err)

	runCmd := "/opt/florist/florist.%s --log-level DEBUG"
	cloudConfig := cloudinit.CloudConfig{
		RunCmd: []string{runCmd},
		WriteFiles: []cloudinit.WriteFile{
			{
				Path:        "/opt/florist/config.json",
				Encoding:    "text/plain",
				Content:     string(floristConfigJson),
				Permissions: "0600",
			},
		},
	}
	rendered, err := cloudConfig.Render()
	rosina.AssertNoError(t, err)
	rosina.AssertTextEqual(t, string(rendered), "ciccio", "rendered")
}
