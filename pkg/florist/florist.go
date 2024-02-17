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
	WorkDir       = "/tmp/florist.work"
	CacheValidity = 24 * time.Hour
)

var currentUser *user.User

func Init() error {
	log := Log.Named("Init")
	var err error
	currentUser, err = user.Current()
	if err != nil {
		return fmt.Errorf("florist.Init: %s", err)
	}
	if err := Mkdir(WorkDir, currentUser.Username, 0755); err != nil {
		return fmt.Errorf("florist.Init: %s", err)
	}
	log.Info("success")
	return nil
}

func User() *user.User {
	if currentUser == nil {
		panic("must call florist.Init before")
	}
	return currentUser
}

// SkipIfNotDisposableHost skips the test if it is running on a precious host.
func SkipIfNotDisposableHost(t *testing.T) {
	t.Helper()
	_, err := os.Stat("/opt/florist/disposable")
	if errors.Is(err, fs.ErrNotExist) {
		t.Skip("skip: this host is not disposable")
	}
}
