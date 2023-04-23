package florist_test

import (
	"testing"

	"gotest.tools/v3/assert"

	"github.com/marco-m/florist/pkg/florist"
)

func TestUserAddSuccessVM(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)

	err := florist.UserAdd("beppe")

	assert.NilError(t, err)
}

func TestSupplementaryGroupsFailureVM(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)

	err := florist.SupplementaryGroups("beppe", "banana")

	assert.ErrorContains(t, err, "group 'banana' does not exist")
}

func TestUserSystemAddSuccessVM(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)

	err := florist.UserSystemAdd("maniglia", "/opt/maniglia")

	assert.NilError(t, err)
}
