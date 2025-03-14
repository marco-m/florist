package florist_test

import (
	"os"
	"os/user"
	"path/filepath"
	"testing"

	"github.com/marco-m/florist/pkg/florist"
	"github.com/marco-m/rosina"
)

func TestMkdirNonExistingDirSuccess(t *testing.T) {
	fPath := filepath.Join(t.TempDir(), "foo")
	owner, group := whoami(t)

	err := florist.Mkdir(fPath, 0o664, owner, group)
	rosina.AssertNoError(t, err)
}

func TestMkdirExistingDirSuccess(t *testing.T) {
	fPath := filepath.Join(t.TempDir(), "foo")
	owner, group := whoami(t)
	perm := os.FileMode(0o664)

	// Make the existing directory, with different permissions.
	err := os.Mkdir(fPath, 0o400)
	rosina.AssertNoError(t, err)

	// SUT
	err = florist.Mkdir(fPath, perm, owner, group)
	rosina.AssertNoError(t, err)

	// Verify SUT has proceeded in the implementation and changed permissions.
	fi, err := os.Stat(fPath)
	rosina.AssertNoError(t, err)
	rosina.AssertEqual(t, fi.Mode().Perm(), perm, "permissions")
}

func TestMkdirFailure(t *testing.T) {
	fPath := filepath.Join(t.TempDir(), "foo")
	owner, group := "non-existing", "non-existing"
	perm := os.FileMode(0o664)

	// Make the existing directory, with different permissions.
	err := os.Mkdir(fPath, 0o400)
	rosina.AssertNoError(t, err)

	// SUT
	err = florist.Mkdir(fPath, perm, owner, group)
	rosina.AssertErrorContains(t, err, "unknown user")
}

// whoami returns the user name and group name of the current user.
func whoami(t *testing.T) (string, string) {
	theUser, err := user.Current()
	rosina.AssertNoError(t, err)
	theGroup, err := user.LookupGroupId(theUser.Gid)
	rosina.AssertNoError(t, err)

	return theUser.Name, theGroup.Name
}
