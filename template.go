package florist

import (
	"fmt"
	"io/fs"
)

// MakeTmplData returns a map, where each key is a path in the secrets FS and each value is
// the contents of the corresponding path.
func MakeTmplData(secrets fs.FS, keys ...string) (map[string]string, error) {
	var m = make(map[string]string, len(keys))
	var errs []error
	for _, k := range keys {
		v, err := fs.ReadFile(secrets, k)
		if err != nil {
			errs = append(errs, err)
		} else {
			m[k] = string(v)
		}
	}

	if len(errs) > 0 {
		matches, err := fs.Glob(secrets, "*")
		if err != nil {
			errs = append(errs, err)
		}
		return m, fmt.Errorf("MakeTmplData (root: %s): %s", matches, errs)
	}
	return m, nil
}
