package florist_test

import (
	"strings"
	"testing"

	"github.com/marco-m/florist"
	"github.com/marco-m/xprog"
)

func TestUserAddSuccessVM(t *testing.T) {
	if xprog.Absent() {
		t.Skip("skip: test requires xprog")
	}

	_, err := florist.UserAdd("beppe")

	if err != nil {
		t.Fatalf("\nhave: %s\nwant: <no error>", err)
	}
}

func TestSupplementaryGroupsFailureVM(t *testing.T) {
	if xprog.Absent() {
		t.Skip("skip: test requires xprog")
	}

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
	if xprog.Absent() {
		t.Skip("skip: test requires xprog")
	}

	_, err := florist.UserSystemAdd("maniglia", "/opt/maniglia")

	if err != nil {
		t.Fatalf("\nhave: %s\nwant: <no error>", err)
	}
}
