// Package florist helps to create non-idempotent, one-file-contains-everything
// installers/provisioners.
package florist

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/user"
	"testing"
	"time"
)

type Flower interface {
	Init() error
	fmt.Stringer
	Description() string
	Install(files fs.FS, finder Finder) error
	Configure(files fs.FS, finder Finder) error
}

const (
	WorkDir       = "/tmp/florist.work"
	CacheValidity = 24 * time.Hour
)

func Init() (*user.User, error) {
	log := Log.Named("Init")

	currUser, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("florist.Init: %s", err)
	}
	if err := Mkdir(WorkDir, currUser.Username, 0755); err != nil {
		return currUser, fmt.Errorf("florist.Init: %s", err)
	}
	log.Info("success")
	return currUser, nil
}

// SkipIfNotDisposableHost skips the test if it is running on a precious host.
func SkipIfNotDisposableHost(t *testing.T) {
	t.Helper()
	_, err := os.Stat("/opt/florist/disposable")
	if errors.Is(err, fs.ErrNotExist) {
		t.Skip("skip: this host is not disposable")
	}
}
