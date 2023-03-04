package florist

import (
	"errors"
	"io/fs"
)

// MakeTmplData returns a map, where each key is a path in fsys and each value is the
// contents of the corresponding path.
func MakeTmplData(fsys fs.FS, keys ...string) (map[string]string, error) {
	var m = make(map[string]string, len(keys))
	var errs []error
	for _, k := range keys {
		v, err := fs.ReadFile(fsys, k)
		if err != nil {
			errs = append(errs, err)
		} else {
			m[k] = string(v)
		}
	}

	return m, errors.Join(errs...)
}
