package cook

import (
	"errors"
	"fmt"
	"os/exec"
	"path"
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
//	env, err := gopassEnv(GopassPrefix+ws,
//		"TF_VAR_consul_gossip_key", "consul_gossip_key",
//		"TF_VAR_hcloud_token",      "hcloud_token"
//	)
func GopassEnv(prefix string, kv ...string) ([]string, error) {
	Out("gopass: reading secrets")
	if len(kv)%2 != 0 {
		return nil, fmt.Errorf("gopassEnv: want an even number of key/vals; have: %v", kv)
	}
	env := make([]string, 0, (len(kv)/2)+1)
	for i := 0; i+1 < len(kv); i += 2 {
		key := kv[i]
		val := kv[i+1]
		// FIXME replace with
		// out, err := GopassGet(path.Join(prefix, val))
		out, err := ExecOut(nil, "gopass", "show", path.Join(prefix, val))
		if err != nil {
			return nil, err
		}
		env = append(env, key+"="+strings.TrimSpace(out))
	}
	return env, nil
}

// GopassDelete removes key.
// A non-existing key is not considered an error. To distinguish the case of a
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
		if strings.Contains(msg, "Secret") && strings.Contains(msg, "does not exist") {
			return nil
		}
		return err
	}
	return nil
}

// GopassGet returns the secret corresponding to key.
// If key does not exist, it returns [ErrNotFound].
// FIXME use gopass cat and stdout
func GopassGet(key string) (string, error) {
	out, err := ExecOut(nil, "gopass", "cat", key)
	if err != nil {
		if strings.Contains(err.Error(), "entry is not in the password store") {
			return "", ErrNotFound
		}
		return "", err
	}
	return out, nil
}

// GopassPut inserts key with associated secret.
// If key already exists, it will be overwritten. To distinguish the case of an existing
// key, use [GopassGet].
// FIXME use gopass cat and stdin
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
