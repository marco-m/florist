// Package florist helps to create non-idempotent, one-file-contains-everything
// installers/provisioners.
package florist

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

const (
	// HomeDir is the directory where to find the default JSON settings file.
	// If not existing, will be created.
	HomeDir = "/opt/florist"
)

// Implementing Flower makes a flower a Flower :-)
type Flower interface {
	Installer
	Configurer
}

// Part of a Flower.
type Installer interface {
	// String is the one-word name of the Flower.
	String() string
	// Description is a one-line description of the Flower.
	Description() string
	// Embedded returns the files embedded in the Flower.
	// See https://pkg.go.dev/embed
	Embedded() []string
	// Init is called at ANY subcommand: list, install, configure.
	// It should store gdn so that it can be used by the other methods.
	Init(gdn *Garden) error
	// Install is called at install time, after Init.
	Install() error
}

// Part of a Flower.
type Configurer interface {
	// Configure is called at configure time, after Init.
	Configure() error
}

// SkipIfNotDisposableHost skips the test if it is running on a precious host.
func SkipIfNotDisposableHost(t *testing.T) {
	t.Helper()
	_, err := os.Stat(filepath.Join(HomeDir, "disposable"))
	if errors.Is(err, fs.ErrNotExist) {
		t.Skip("skip: this host is not disposable")
	}
}

// Ptr returns a pointer to 'p'.
// Needed to fill a struct field of primitive type (int, string and similar) when the Go
// zero value cannot be used because it is a valid value for the field.
//
// Deprecated: Since Go 1.26, this function is not needed. Call new(p) directly.
// If you run "go fix" on your codebase, this function will be inlined for you.
//
//go:fix inline
func Ptr[T any](p T) *T {
	return new(p)
}
