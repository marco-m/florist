package florist

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/user"
	"strconv"
)

// Create directory path with associated owner and permissions.
// If the directory already exists it does nothing.
func Mkdir(path string, owner *user.User, perm fs.FileMode) error {
	log := Log.Named("Mkdir").With("path", path)

	uid, _ := strconv.Atoi(owner.Uid)
	gid, _ := strconv.Atoi(owner.Gid)

	_, err := os.Stat(path)
	if err == nil {
		log.Debug("directory exists")
		return nil
	}
	if !errors.Is(err, fs.ErrNotExist) {
		// Some other error
		return fmt.Errorf("florist.Mkdir: %s", err)
	}
	// Directory doesn't exist, let's create it
	if err := os.Mkdir(path, perm); err != nil {
		return fmt.Errorf("florist.Mkdir: %s", err)
	}
	log.Debug("directory created", "perm", perm)
	if err := os.Chown(path, uid, gid); err != nil {
		return fmt.Errorf("florist.Mkdir: %s", err)
	}
	log.Debug("directory chown", "owner", owner.Username)
	return nil
}
