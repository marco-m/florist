package florist_test

import (
	"os"
	"os/user"
	"path/filepath"
	"testing"

	"github.com/marco-m/florist/pkg/florist"
	"github.com/marco-m/rosina/assert"
)

func TestMkdirNonExistingDirSuccess(t *testing.T) {
	fPath := filepath.Join(t.TempDir(), "foo")
	owner, group := whoami(t)

	err := florist.Mkdir(fPath, 0o664, owner, group)
	assert.NoError(t, err, "florist.Mkdir")
}

func TestMkdirExistingDirSuccess(t *testing.T) {
	fPath := filepath.Join(t.TempDir(), "foo")
	owner, group := whoami(t)
	perm := os.FileMode(0o664)

	// Make the existing directory, with different permissions.
	err := os.Mkdir(fPath, 0o400)
	assert.NoError(t, err, "os.Mkdir")

	// SUT
	err = florist.Mkdir(fPath, perm, owner, group)
	assert.NoError(t, err, "florist.Mkdir")

	// Verify SUT has proceeded in the implementation and changed permissions.
	fi, err := os.Stat(fPath)
	assert.NoError(t, err, "os.Stat")
	assert.Equal(t, fi.Mode().Perm(), perm, "permissions")
}

func TestMkdirFailure(t *testing.T) {
	fPath := filepath.Join(t.TempDir(), "foo")
	owner, group := "non-existing", "non-existing"
	perm := os.FileMode(0o664)

	// Make the existing directory, with different permissions.
	err := os.Mkdir(fPath, 0o400)
	assert.NoError(t, err, "os.Mkdir")

	// SUT
	err = florist.Mkdir(fPath, perm, owner, group)
	assert.ErrorContains(t, err, "unknown user")
}

// whoami returns the user name and group name of the current user.
func whoami(t *testing.T) (string, string) {
	theUser, err := user.Current()
	assert.NoError(t, err, "user.Current")
	theGroup, err := user.LookupGroupId(theUser.Gid)
	assert.NoError(t, err, "user.LookupGroupId")

	return theUser.Name, theGroup.Name
}
