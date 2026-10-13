// Package florist helps to create non-idempotent, one-file-contains-everything
// installers/provisioners.
package florist

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
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

// Treat this as a constant. Override only for testing of Florist itself.
// Unexported on purpose.
var kDisposableMarkerPath = "/florist.disposable"

// Use only for testing.
func IsDisposableHost() bool {
	_, err := os.Stat(kDisposableMarkerPath)
	if err == nil {
		// Expected when file exists.
		return true
	}
	if errors.Is(err, fs.ErrNotExist) {
		// Expected when file does not exist.
		return false
	}
	// Unexpected other error. Fail-safe and report this as NOT disposable!
	slog.Info("florist.IsDisposableHost", "error", err)
	return false
}

// Use only for testing.
func SetDisposableHost() error {
	return fmt.Errorf("florist.SetDisposableHost: not implemented")
}
