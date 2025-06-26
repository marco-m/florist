package florist_test

import (
	"fmt"
	"math/rand/v2"
	"path"
	"strconv"
	"strings"
	"testing"

	"github.com/marco-m/florist/pkg/florist"
)

func TestUserAddSuccess(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)

	// Ensure user doesn't exist.
	name := "florist-" + strconv.Itoa(rand.IntN(100_000)+200_000)

	err1 := florist.UserAdd(name, nil)
	if err1 != nil {
		t.Errorf("\nadding non-existing user:\nhave error: %s\nwant: <no error>", err1)
	}

	// Now the user is already present, this should not be an error.
	err2 := florist.UserAdd(name, nil)
	if err2 != nil {
		t.Errorf("\nadding existing user:\nhave error: %s\nwant: <no error>", err1)
	}

	// The home directory has been created.
	homeDir := path.Join("/home", name)
	exists, err := florist.FileExists(homeDir)
	if !exists {
		t.Error("home directory does not exist:", homeDir)
	}
	if err != nil {
		t.Errorf("\nwhile checking for existence of %s: %s", homeDir, err)
	}
}

func TestUserAddSystemSuccess(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)

	// Ensure user doesn't exist.
	name := "florist-" + strconv.Itoa(rand.IntN(100_000)+200_000)
	homeDir := path.Join("/opt", name)

	err := florist.UserAdd(name, &florist.UserAddArgs{

		System:  true,
		HomeDir: homeDir,
	})
	if err != nil {
		t.Errorf("\nhave error: %s\nwant: <no error>", err)
	}

	// The home directory has been created.
	exists, err := florist.FileExists(homeDir)
	if !exists {
		t.Error("home directory does not exist:", homeDir)
	}
	if err != nil {
		t.Errorf("\nwhile checking for existence of %s: %s", homeDir, err)
	}
}

func TestUserAddWithSupplementaryGroupsSuccess(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)

	// Ensure user doesn't exist.
	name := "florist-" + strconv.Itoa(rand.IntN(100_000)+200_000)

	group := "banana"
	err := florist.GroupAdd(group, nil)
	if err != nil {
		t.Errorf("\nhave error: %s\nwant: <no error>", err)
	}

	err = florist.UserAdd(name, &florist.UserAddArgs{
		Groups: []string{group},
	})
	if err != nil {
		t.Errorf("\nhave error: %s\nwant: <no error>", err)
	}
}

func TestUserAddWithSupplementaryGroupsFailure(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)

	// Ensure user doesn't exist.
	name := "florist-" + strconv.Itoa(rand.IntN(100_000)+200_000)

	// Ensure group doesn't exist.
	group := "group-" + strconv.Itoa(rand.IntN(100_000)+200_000)

	err := florist.UserAdd(name, &florist.UserAddArgs{
		Groups: []string{group},
	})
	if err == nil {
		t.Fatalf("\nhave: <no error>\nwant: non-nil error")
	}
	have := err.Error()
	needle := fmt.Sprintf("group '%s' does not exist", group)
	if !strings.Contains(have, needle) {
		t.Errorf("\nerror message:    %s\ndoes not contain: %s", have, needle)
	}
}
