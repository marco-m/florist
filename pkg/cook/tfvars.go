package cook

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
)

// ParseTfVars takes a tfVars HCL file and returns a map containing the k/v.
// NOTE: Type information is NOT preserved: everything is a string.
// Terraform identifiers can contain letters, digits, underscores (_)
// and hyphens (-).
func ParseTfVars(tfVarsFile string) (map[string]string, error) {
	fi, err := os.Open(tfVarsFile)
	if err != nil {
		return nil, fmt.Errorf("ParseTfVars: %s", err)
	}
	return parseTfVars(fi)
}

var kvRe = regexp.MustCompile(`^([a-zA-Z]+[-_0-9a-zA-z]*)\s+=\s+"*([-\w]+)"*$`)

func parseTfVars(rd io.Reader) (map[string]string, error) {
	kv := make(map[string]string)
	scanner := bufio.NewScanner(rd)
	for scanner.Scan() {
		matches := kvRe.FindStringSubmatch(scanner.Text())
		if matches != nil {
			k, v := matches[1], matches[2]
			if _, ok := kv[k]; ok {
				return nil, fmt.Errorf("parseTfVars: duplicate key: %s", k)
			}
			kv[k] = v
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return kv, nil
}

// TfVarsToDir takes a tfVars HCL file and puts each k/v as a file directly
// below directory dir (flat).
// NOTE: Type information is NOT preserved: everything is a string.
func TfVarsToDir(tfVarsFile string, dir string) error {
	kv, err := ParseTfVars(tfVarsFile)
	if err != nil {
		return fmt.Errorf("TfVarsToDir: %s", err)
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("TfVarsToDir: %s", err)
	}
	for k, v := range kv {
		if err := os.WriteFile(filepath.Join(dir, k), []byte(v), 0640); err != nil {
			return fmt.Errorf("TfVarsToDir: %s", err)
		}
	}

	return nil
}
