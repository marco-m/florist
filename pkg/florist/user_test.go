package florist_test

import (
	"strings"
	"testing"

	"github.com/marco-m/florist/pkg/florist"
)

func TestUserAddSuccess(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)

	err := florist.UserAdd("beppe")
	if err != nil {
		t.Errorf("\nhave error: %s\nwant: <no error>", err)
	}
}

func TestSupplementaryGroupsFailure(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)

	err := florist.SupplementaryGroups("beppe", "banana")
	if err == nil {
		t.Fatalf("\nhave: <no error>\nwant: non-nil error")
	}
	have := err.Error()
	needle := "group 'banana' does not exist"
	if !strings.Contains(have, needle) {
		t.Errorf("\nerror message:    %s\ndoes not contain: %s", have, needle)
	}
}

func TestUserSystemAddSuccess(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)

	err := florist.UserSystemAdd("maniglia", "/opt/maniglia")
	if err != nil {
		t.Errorf("\nhave error: %s\nwant: <no error>", err)
	}
}
