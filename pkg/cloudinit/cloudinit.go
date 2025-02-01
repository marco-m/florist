// Package cloudinit is a simple approach to generate a cloud-config file
// for cloud-init. To avoid bringing in a YAML library, it exploits the fact
// that YAML is a superset of JSON, so any JSON object is valid YAML.
package cloudinit

import (
	"encoding/json"
	"fmt"

	"github.com/creasty/defaults"
)

// CloudConfig represents the cloud-config file. Fill the fields that you need
// and call [CloudConfig.Render].
//
// The field names and descriptions are from the cloud-config documentation [1].
// Note that all the unfilled fields will have the default value specified by
// the cloud-config documentation [1], not their Go zero value (for example, it
// is possible that a boolean field gets true as default value. Read the field
// documentation in case of doubt).
//
// Only the subset of fields that I need are available. Additional fields from
// upstream could be added.
//
// To aid in debugging, cloud-init logs everything to
// /var/log/cloud-init-output.log, including the output of [CloudConfig.RunCmd].
//
// [1]: https://cloudinit.readthedocs.io/en/latest/explanation/about-cloud-config.html
type CloudConfig struct {
	// Run arbitrary commands at a rc.local-like time-frame with output to the
	// console. The items will be executed as if passed to execve(3) (with the
	// first argument as the command). The 'runcmd' module only writes the
	// script to be run later. The module that actually runs the script is
	// 'scripts_user' in the Final boot stage.
	// https://cloudinit.readthedocs.io/en/latest/reference/modules.html#runcmd
	RunCmd []string `json:"runcmd,omitzero"`

	//
	// SSH
	// https://cloudinit.readthedocs.io/en/latest/reference/modules.html#ssh
	//
	// Remove host SSH keys. This prevents re-use of a private host key from an
	// image with default host SSH keys. Default: true.
	SshDeleteKeys bool `json:"ssh_deletekeys,omitzero"`
	// The SSH key types to generate. Default: [rsa, ecdsa, ed25519].
	SshGenKeyTypes []string `json:"ssh_genkeytypes,omitzero"`

	// Write arbitrary files.
	// https://cloudinit.readthedocs.io/en/latest/reference/modules.html#write-files
	WriteFiles []WriteFile `json:"write_files,omitzero"`
}

// Render returns the cloud-config object as JSON (which is valid YAML).
//
// NOTE The output of Render can still be an invalid cloud-config file, exactly
// as you could write by hand a syntactically valid YAML file that is an invalid
// cloud-config file.
func (cfg *CloudConfig) Render() ([]byte, error) {
	if err := defaults.Set(cfg); err != nil {
		return nil, fmt.Errorf("render: applying defaults: %s", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return nil, err
	}
	return append([]byte("#cloud-config\n"), data...), nil
}

// Write a file.
// https://cloudinit.readthedocs.io/en/latest/reference/modules.html#write-files
type WriteFile struct {
	// Path of the file to which Content is decoded and written.
	Path string `json:"path"`
	// Optional content to write to the provided Path. When content is present
	// and encoding is not ‘text/plain’, decode the content prior to writing.
	Content string `json:"content,omitzero"`
	//  Optional owner:group to chown on the file and new directories. Default:
	//  root:root.
	Owner string `json:"owner,omitzero"`
	// Optional file permissions to set on Path represented as an octal string
	// ‘0###’. Default: 0o644.
	Permissions string `json:"permissions,omitzero"`
	// Optional encoding type of the content. Default: text/plain. Supported
	// encoding types are: gz, gzip, gz+base64, gzip+base64, gz+b64, gzip+b64,
	// b64, base64.
	Encoding string `json:"encoding,omitzero"`
}
