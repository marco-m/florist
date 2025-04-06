package florist_test

import (
	"testing"

	"github.com/marco-m/florist/pkg/florist"
	"github.com/marco-m/rosina/assert"
)

func TestUserAddSuccessVM(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)

	err := florist.UserAdd("beppe")

	assert.NoError(t, err, "florist.UserAdd")
}

func TestSupplementaryGroupsFailureVM(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)

	err := florist.SupplementaryGroups("beppe", "banana")

	assert.ErrorContains(t, err, "group 'banana' does not exist")
}

func TestUserSystemAddSuccessVM(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)

	err := florist.UserSystemAdd("maniglia", "/opt/maniglia")

	assert.NoError(t, err, "florist.UserSystemAdd")
}
