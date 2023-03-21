package florist

import (
	"fmt"
	"io/fs"
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

func (uf *UnionFinder) Error() error {
	return uf.finders[len(uf.finders)-1].Error()
}
