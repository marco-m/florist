// Florist helps creating non idempotent, one-file-contains-everything
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

// A Flower is a composable unit that can be installed.
type Flower interface {
	fmt.Stringer
	Description() string
	Init() error
	Install() error
	Configure() error
}

const (
	WorkDir              = "/tmp/florist.work"
	CacheValidityDefault = 24 * time.Hour
)

func Init() (*user.User, error) {
	log := Log.Named("Init")

	currUser, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("florist.Init: %s", err)
	}
	if err := Mkdir(WorkDir, currUser, 0755); err != nil {
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
