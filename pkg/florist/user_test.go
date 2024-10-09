package florist_test

import (
	"testing"

	"github.com/marco-m/florist/pkg/florist"
	"github.com/marco-m/rosina"
)

func TestUserAddSuccessVM(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)

	err := florist.UserAdd("beppe")

	rosina.AssertNoError(t, err)
}

func TestSupplementaryGroupsFailureVM(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)

	err := florist.SupplementaryGroups("beppe", "banana")

	rosina.AssertErrorContains(t, err, "group 'banana' does not exist")
}

func TestUserSystemAddSuccessVM(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)

	err := florist.UserSystemAdd("maniglia", "/opt/maniglia")

	rosina.AssertNoError(t, err)
}
