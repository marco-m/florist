package florist_test

import (
	"testing"

	"github.com/marco-m/florist/pkg/florist"
	"github.com/marco-m/rosina"
)

func TestUserAddSuccessVM(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)

	err := florist.UserAdd("beppe")

	rosina.AssertIsNil(t, err)
}

func TestSupplementaryGroupsFailureVM(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)

	err := florist.SupplementaryGroups("beppe", "banana")

	rosina.AssertIsNotNil(t, err)
	rosina.AssertContains(t, err.Error(), "group 'banana' does not exist")
}

func TestUserSystemAddSuccessVM(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)

	err := florist.UserSystemAdd("maniglia", "/opt/maniglia")

	rosina.AssertIsNil(t, err)
}
