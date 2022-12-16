package florist

import (
	"errors"
	"io/fs"
)

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
