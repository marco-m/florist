// Package cloudinit is a simple approach to generate a cloud-config file
// for cloudinit. To avoid bringing in a YAML library, it exploits the fact
// that YAML is a superset of JSON, so any JSON object is valid YAML.
package cloudinit

import (
	"encoding/json"
	"fmt"

	"github.com/creasty/defaults"
)

// The defaults and the field descriptions are taken from
// https://cloudinit.readthedocs.io
type CloudConfig struct {
	// Run arbitrary commands at a rc.local-like time-frame with output to the
	// console. The items will be executed as if passed to execve(3) (with the
	// first argument as the command). The runcmd module only writes the script
	// to be run later. The module that actually runs the script is scripts_user
	// in the Final boot stage.
	RunCmd []string `json:"runcmd,omitzero"`
	// Remove host SSH keys. This prevents re-use of a private host key from an
	// image with default host SSH keys. Default: true.
	SshDeleteKeys bool `json:"ssh_deletekeys,omitzero" default:"true"`
	// The SSH key types to generate. Default: [rsa, ecdsa, ed25519].
	SshGenKeyTypes []string `json:"ssh_genkeytypes,omitzero" default:"[\"rsa\", \"ecdsa\", \"ed25519\"]"`
	// Write arbitrary files.
	WriteFiles []WriteFile `json:"write_files,omitzero"`
}

// Render returns the cloud-config object as JSON (which is valid YAML).
func (cfg *CloudConfig) Render() ([]byte, error) {
	if err := defaults.Set(cfg); err != nil {
		return nil, fmt.Errorf("render: applying defaults: %s", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return nil, err
	}
	return append([]byte("#cloud-config\n\n"), data...), nil
}

type WriteFile struct {
	// Path of the file to which Content is decoded and written.
	Path string `json:"path"`
	// Optional content to write to the provided Path. When content is present
	// and encoding is not ‘text/plain’, decode the content prior to writing.
	// Default: ''.
	Content string `json:"content,omitzero"`
	//  Optional owner:group to chown on the file and new directories. Default:
	//  root:root.
	Owner string `json:"owner,omitzero"`
	// Optional file permissions to set on Path represented as an octal string
	// ‘0###’. Default: 0o644.
	Permissions string `json:"permissions,omitzero"`
	// Optional encoding type of the content. Default: text/plain. No decoding
	// is performed by default. Supported encoding types are: gz, gzip,
	// gz+base64, gzip+base64, gz+b64, gzip+b64, b64, base64.
	Encoding string `json:"encoding"`
}
