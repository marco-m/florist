package cook

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

var ErrNotFound = errors.New("not found")

// GopassEnv returns a slice ready to be used as environment variables to be passed to
// a subprocess. The slice uses as keys the odd-numbered parameters (the "k") and as
// values the gopass secrets corresponding to the even-numbered parameters (the "v").
// Parameter "prefix" is prefixed to each "v".
// It returns an error if there is an error during gopass operations, for example if a
// "v" does not exist in the gopass store or if it cannot find the gopass executable.
//
// Example:
//
//	env, err := GopassEnv(GopassPrefix+ws,
//		"TF_VAR_hcloud_token",      "hcloud_token"
//	)
//
// See also [GopassToConfig].
func GopassEnv(prefix string, kv ...string) ([]string, error) {
	Out("gopass: reading secrets to env")
	if len(kv)%2 != 0 {
		return nil, fmt.Errorf("GopassEnv: want an even number of key/vals; have: %v", kv)
	}
	env := make([]string, 0, (len(kv)/2)+1)
	for i := 0; i+1 < len(kv); i += 2 {
		key := kv[i]
		val := kv[i+1]
		path := path.Join(prefix, val)
		out, err := GopassGet(path)
		if err != nil {
			return nil, fmt.Errorf("GopassEnv: looking for %s: %s", path, err)
		}
		env = append(env, key+"="+strings.TrimSpace(out))
	}
	return env, nil
}

// GopassToConfig creates file secretsPath, containing one JSON object, whose keys are
// the odd-numbered parameters (the "k") and whose values are the gopass secrets
// corresponding to the even-numbered parameters (the "v").
// Parameter "prefix" is prefixed to each "v".
//
// Example:
//
//	err := GopassToConfig("hcloud/prod/controllers", "florist.controllers.config.json",
//		"ssh_host_ed25519_key",          "ssh_host_ed25519_key",
//		"ssh_host_ed25519_key_cert_pub", "ssh_host_ed25519_key-cert.pub",
//	)
//
// See also [GopassEnv].
func GopassToConfig(secretsPath string, prefix string, kv ...string) error {
	Out("gopass: reading secrets to file")
	if len(kv)%2 != 0 {
		return fmt.Errorf("GopassToConfig: want an even number of key/vals; have: %v", kv)
	}

	fi, err := os.OpenFile(secretsPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("GopassToConfig: %s", err)
	}
	defer fi.Close()

	m := make(map[string]string, len(kv)/2)
	for i := 0; i+1 < len(kv); i += 2 {
		key := kv[i]
		val := kv[i+1]
		path := path.Join(prefix, val)
		out, err := GopassGet(path)
		if err != nil {
			return fmt.Errorf("GopassToConfig: looking for %s: %s", path, err)
		}
		m[key] = out
	}

	enc := json.NewEncoder(fi)
	enc.SetIndent("", "  ")
	if err := enc.Encode(m); err != nil {
		return fmt.Errorf("GopassToConfig: %s", err)
	}

	return nil
}

// GopassToDir decrypts and exports to dstDir all the gopass secrets below srdDir.
// If dstDir does not exist, GopassToDir will create it.
// NOTE It is up to the caller to delete dstDir, the directory with exported secrets.
func GopassToDir(srcDir string, dstDir string) error {
	keys, err := GopassLs(srcDir)
	if err != nil {
		return fmt.Errorf("GopassToDir: %s", err)
	}
	if len(keys) == 0 {
		return fmt.Errorf("GopassToDir: no keys below %s", srcDir)
	}

	seen := make(map[string]bool)
	for _, key := range keys {
		dstFile := filepath.Join(dstDir, key)
		// "a/b/k3" => "a/b", "k3"
		dir, _ := path.Split(dstFile)
		if dir != "" && !seen[dir] {
			seen[dir] = true
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("GopassToDir: %s", err)
			}
		}
		secret, err := GopassGet(path.Join(srcDir, key))
		if err != nil {
			return fmt.Errorf("GopassToDir: %s", err)
		}
		f, err := os.OpenFile(dstFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			return fmt.Errorf("GopassToDir: %s", err)
		}
		_, err = f.WriteString(secret)
		if err != nil {
			return fmt.Errorf("GopassToDir: %s", err)
		}
	}

	return nil
}

// GopassDelete removes key.
// Secret non-existing key is not considered an error. To distinguish the case of a
// non-existent key, use [GopassGet].
func GopassDelete(key string) error {
	// --force is needed to avoid getting a prompt, but if the key doesn't exist,
	// it still returns an error, for example:
	//   gopass delete --force cook/foo
	//   (yes empty line...)
	//   Error: Secret "cook/foo" does not exist
	_, err := ExecOut(nil, "gopass", "delete", "--force", key)
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "does not exist") ||
			strings.Contains(msg, "entry is not in the password store") {
			return nil
		}
		return err
	}
	return nil
}

// GopassGet returns the secret corresponding to key.
// If key does not exist, it returns [ErrNotFound].
func GopassGet(key string) (string, error) {
	out, err := ExecOut(nil, "gopass", "cat", key)
	if err != nil {
		if strings.Contains(err.Error(), "entry is not in the password store") {
			return "", ErrNotFound
		}
		return "", err
	}
	// Sigh.
	// return strings.TrimSpace(out), nil
	return out, nil
}

// GopassPut inserts key with associated secret.
// If key already exists, it will be overwritten. To distinguish the case of an existing
// key, use [GopassGet].
func GopassPut(key string, secret string) error {
	cmd := exec.Command("gopass", "cat", key)
	cmd.Stdin = strings.NewReader(secret)
	out, err := cmd.Output()
	if err != nil {
		return err
	}
	if len(out) > 0 {
		return fmt.Errorf("unexpected: %s", string(out))
	}
	return nil
}

// GopassLs returns a list of the keys below prefix, with the prefix stripped.
func GopassLs(prefix string) ([]string, error) {
	var keys []string

	out, err := ExecOut(nil, "gopass", "ls", "--flat", "--strip-prefix", prefix)
	if err != nil {
		return keys, fmt.Errorf("GopassLs: %s", err)
	}
	keys = append(keys, strings.Fields(out)...)

	return keys, nil
}
