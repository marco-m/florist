package florist_test

import (
	"os/user"
	"strconv"
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

func TestUserSystemAddWithUIDSuccessVM(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)

	userName := "giuseppe"
	userUid := 899
	err := florist.UserSystemAddWithUID(userName, "/opt/giuseppe", userUid)
	assert.NoError(t, err, "florist.UserSystemAddWithUID")

	haveUser, err := user.Lookup(userName)
	assert.NoError(t, err, "florist.user.Lookup")

	haveUid, err := strconv.Atoi(haveUser.Uid)
	assert.NoError(t, err, "strconv.Atoi")

	assert.Equal(t, haveUid, userUid, "user uid")
}
