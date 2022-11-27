package florist_test

import (
	"strings"
	"testing"

	"github.com/marco-m/florist"
)

func TestUserAddSuccessVM(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)

	_, err := florist.UserAdd("beppe")

	if err != nil {
		t.Fatalf("\nhave: %s\nwant: <no error>", err)
	}
}

func TestSupplementaryGroupsFailureVM(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)

	err := florist.SupplementaryGroups("beppe", "banana")

	wantErr := "group 'banana' does not exist"
	if err == nil {
		t.Fatalf("\nhave: <no error>\nwant: %s", wantErr)
	}
	have := err.Error()
	if !strings.Contains(have, wantErr) {
		t.Fatalf("\nhave: %s\nwant: contains: %s", have, wantErr)
	}
}

func TestUserSystemAddSuccessVM(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)

	_, err := florist.UserSystemAdd("maniglia", "/opt/maniglia")

	if err != nil {
		t.Fatalf("\nhave: %s\nwant: <no error>", err)
	}
}
