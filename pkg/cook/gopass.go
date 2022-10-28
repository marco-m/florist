package cook

import (
	"fmt"
	"path"
	"strings"
)

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
		out, err := ExecOut(nil, "gopass", "show", path.Join(prefix, val))
		if err != nil {
			return nil, err
		}
		env = append(env, key+"="+strings.TrimSpace(out))
	}
	return env, nil
}
