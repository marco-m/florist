package florist_test

import (
	"math/rand/v2"
	"strconv"
	"strings"
	"testing"

	"github.com/marco-m/florist/pkg/florist"
)

func TestUserAddSuccess(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)

	// Ensure user doesn't exist.
	name := "florist-" + strconv.Itoa(rand.IntN(100_000)+200_000)

	err1 := florist.UserAdd(name, florist.UserAddOpt{})
	if err1 != nil {
		t.Errorf("\nadding non-existing user:\nhave error: %s\nwant: <no error>", err1)
	}

	err2 := florist.UserAdd(name, florist.UserAddOpt{})
	if err2 != nil {
		t.Errorf("\nadding existing user:\nhave error: %s\nwant: <no error>", err1)
	}
}

func TestUserAddSystemSuccess(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)

	err := florist.UserAdd("maniglia", florist.UserAddOpt{
		System:  true,
		HomeDir: "/opt/maniglia",
	})
	if err != nil {
		t.Errorf("\nhave error: %s\nwant: <no error>", err)
	}
}

func TestSupplementaryGroupsFailure(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)

	// Ensure user doesn't exist.
	name := "florist-" + strconv.Itoa(rand.IntN(100_000)+200_000)

	err := florist.UserAdd(name, florist.UserAddOpt{
		SupplementaryGroups: []string{"banana"},
	})
	if err == nil {
		t.Fatalf("\nhave: <no error>\nwant: non-nil error")
	}
	have := err.Error()
	needle := "group 'banana' does not exist"
	if !strings.Contains(have, needle) {
		t.Errorf("\nerror message:    %s\ndoes not contain: %s", have, needle)
	}
}
