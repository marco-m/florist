package cloudinit_test

import (
	"encoding/json"
	"testing"

	"github.com/marco-m/florist/pkg/cloudinit"
	"github.com/marco-m/rosina/assert"
)

func TestCloudConfigRender(t *testing.T) {
	want := `#cloud-config
{
  "runcmd": [
    "/path/foo bar --inner /path/inner.json"
  ],
  "write_files": [
    {
      "path": "/path/inner.json",
      "content": "{\n  \"Pizza\": \"Margherita\",\n  \"Topping\": \"Pineapple\"\n}",
      "permissions": "0600"
    }
  ]
}`

	type Inner struct {
		Pizza   string
		Topping string
	}
	inner := Inner{
		Pizza:   "Margherita",
		Topping: "Pineapple",
	}

	// MarshalIndent adds newlines, so that the inner file will be written
	// by cloud-init with newlines. No need for additional escaping.
	innerJson, err := json.MarshalIndent(inner, "", "  ")
	assert.NoError(t, err, "JSON encoding inner")

	runCmd := "/path/foo bar --inner /path/inner.json"
	cloudConfig := cloudinit.CloudConfig{
		RunCmd: []string{runCmd},
		WriteFiles: []cloudinit.WriteFile{
			{
				Path:        "/path/inner.json",
				Content:     string(innerJson),
				Permissions: "0600",
			},
		},
	}
	rendered, err := cloudConfig.Render()
	assert.NoError(t, err, "cloudConfig.Render")
	assert.Equal(t, string(rendered), want, "rendered")
}
