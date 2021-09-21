package ssh

import (
	"errors"
	"io/fs"
	"os"
	"os/user"
	"path/filepath"
	"strconv"

	"github.com/marco-m/florist"
)

// This overwrites !
func AddAuthorizedKeys(owner *user.User, vfs fs.FS, srcPath string) error {
	log := florist.Log.Named("ssh.AddAuthorizedKeys").With("user", owner.Username)

	uid, _ := strconv.Atoi(owner.Uid)
	gid, _ := strconv.Atoi(owner.Gid)

	// If $HOME/.ssh doesn't exist, create it, correct owner and permissions.
	sshDir := filepath.Join(owner.HomeDir, ".ssh")
	_, err := os.Stat(sshDir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			if err := os.Mkdir(sshDir, 0700); err != nil {
				return err
			}
			if err := os.Chown(sshDir, uid, gid); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	authorizedKeysPath := filepath.Join(sshDir, "authorized_keys")
	if err := florist.Copy(vfs, srcPath, authorizedKeysPath, 0400, owner); err != nil {
		return err
	}

	log.Debug("added file", "path", authorizedKeysPath)
	return nil
}
