// Package florist helps to create non-idempotent, one-file-contains-everything
// installers/provisioners.
package florist

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
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

// currentUser is set by Init.
var currentUser *user.User

// Log is the logger for the florist module.
// It logs nothing by default. Use Init to change this.
var Log = slog.New(slog.NewTextHandler(io.Discard, nil))

// Init initializes the florist package. Call it as first thing.
// The logger log with be set as the florist logger. If nil, florist will not log.
// We suggest to log to os.Stdout instead of os.Stderr, because Packer renders
// stderr in red, with the effect of making it seem a long stream of errors.
//
// Example:
//
//	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
//		Level: slog.LevelDebug,
//	}))
//	if err := Init(log); err != nil { ...
func Init(log *slog.Logger) error {
	if currentUser != nil {
		Log.Warn("init-already-called")
		return nil
	}
	Log = log.With("lib", "florist")
	var err error
	currentUser, err = user.Current()
	if err != nil {
		return fmt.Errorf("florist.Init: %s", err)
	}
	// FIXME this should be postponed and executed by the components needing it?
	if err := Mkdir(WorkDir, currentUser.Username, 0755); err != nil {
		return fmt.Errorf("florist.Init: %s", err)
	}
	log.Info("init", "status", "success")
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
