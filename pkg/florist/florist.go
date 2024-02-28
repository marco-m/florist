// Package florist helps to create non-idempotent, one-file-contains-everything
// installers/provisioners.
package florist

import (
	"errors"
	"io/fs"
	"os"
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
	WorkDir = "/tmp/florist.work"
)

// SkipIfNotDisposableHost skips the test if it is running on a precious host.
func SkipIfNotDisposableHost(t *testing.T) {
	t.Helper()
	_, err := os.Stat("/opt/florist/disposable")
	if errors.Is(err, fs.ErrNotExist) {
		t.Skip("skip: this host is not disposable")
	}
}
