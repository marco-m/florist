// Package florist helps to create non-idempotent, one-file-contains-everything
// installers/provisioners.
package florist

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

type Flower interface {
	Installer
	Configurer
}

type Installer interface {
	String() string
	Description() string
	Embedded() []string
	Init() error
	Install() error
}

type Configurer interface {
	Configure() error
}

const (
	WorkDir = "/tmp/florist"
	HomeDir = "/opt/florist"
)

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
func Ptr[T any](p T) *T {
	return &p
}

func makeErrorf(prefix string) func(format string, a ...any) error {
	return func(format string, a ...any) error {
		return fmt.Errorf(prefix+": "+format, a...)
	}
}
