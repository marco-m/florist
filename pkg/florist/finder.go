package florist

import (
	"fmt"
	"io/fs"
	"sort"

	"tailscale.com/util/multierr"
)

// Finder searches for keys.
type Finder interface {
	// Get returns the value of key if it exists. If it does not exist, Get
	// records the error and returns the zero value. See also [Error].
	Get(key string) string

	// Lookup returns the value of key if it exists. If it does not exist,
	// Lookup records the error and returns the zero value and an error.
	// See also [Error].
	Lookup(key string) (string, error)

	// Keys returns all the _static_ keys in the finder. This method is useful
	// only for debugging.
	// NOTE Depending on the implementation, it might be impossible to know
	// the keys in advance! In such case, finder will return
	Keys() ([]string, error)

	// Error returns the errors accumulated by all the calls to [Get] and [Lookup].
	Error() error
}

// FsFinder is a [Finder] over a fs.FS filesystem, typically an embed.FS.
type FsFinder struct {
	fsys   fs.FS
	errors []error
}

var _ Finder = (*FsFinder)(nil)

func NewFsFinder(fsys fs.FS) *FsFinder {
	return &FsFinder{
		fsys:   fsys,
		errors: []error{},
	}
}

func (ff *FsFinder) Get(key string) string {
	val, _ := ff.Lookup(key)
	return val
}

func (ff *FsFinder) Lookup(key string) (string, error) {
	buf, err := fs.ReadFile(ff.fsys, key)
	if err != nil {
		err := fmt.Errorf("FsFinder.Get: %s", err)
		ff.errors = append(ff.errors, err)
		return "", err
	}
	return string(buf), nil
}

func (ff *FsFinder) Keys() ([]string, error) {
	keys := []string{}
	fn := func(dePath string, de fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !de.IsDir() {
			keys = append(keys, dePath)
		}
		return nil
	}
	// We ignore errors
	fs.WalkDir(ff.fsys, ".", fn)

	sort.Strings(keys)
	return keys, nil
}

func (ff *FsFinder) Error() error {
	return multierr.New(ff.errors...)
}

// UnionFinder is a [Finder] over multiple Finder instances, like a union mount.
type UnionFinder struct {
	finders []Finder
}

var _ Finder = (*UnionFinder)(nil)

// NewUnionFinder returns an [UnionFinder] made of finders.
// The search for a key starts at the first [Finder] in finders and stops as
// soon as a value is found. Said in another way, finders is sorted in priority
// order.
func NewUnionFinder(finders ...Finder) *UnionFinder {
	return &UnionFinder{
		finders: finders,
	}
}

func (uf *UnionFinder) Get(key string) string {
	val, _ := uf.Lookup(key)
	return val
}

func (uf *UnionFinder) Lookup(key string) (string, error) {
	var err error
	var val string
	for _, fnd := range uf.finders {
		val, err = fnd.Lookup(key)
		if err == nil {
			return val, nil
		}
	}
	return "", err
}

func (uf *UnionFinder) Keys() ([]string, error) {
	seen := make(map[string]struct{})
	for _, fnd := range uf.finders {
		keys, err := fnd.Keys()
		if err != nil {
			return nil, fmt.Errorf("UnionFinder.Keys: %s", err)
		}
		// Deduplicate.
		for _, k := range keys {
			seen[k] = struct{}{}
		}
	}

	uniqueKeys := make([]string, 0, len(seen))
	for k := range seen {
		uniqueKeys = append(uniqueKeys, k)
	}

	sort.Strings(uniqueKeys)
	return uniqueKeys, nil
}

func (uf *UnionFinder) Error() error {
	return uf.finders[len(uf.finders)-1].Error()
}
