package ssh

import (
	"io/fs"
	"os/user"
	"path/filepath"

	"github.com/marco-m/florist"
)

// This overwrites !
func AddAuthorizedKeys(owner *user.User, vfs fs.FS, srcPath string) error {
	log := florist.Log.Named("ssh.AddAuthorizedKeys").With("user", owner.Username)

	// If $HOME/.ssh doesn't exist, create it, correct owner and permissions.
	sshDir := filepath.Join(owner.HomeDir, ".ssh")
	if err := florist.Mkdir(sshDir, owner, 0700); err != nil {
		return err
	}

	authorizedKeysPath := filepath.Join(sshDir, "authorized_keys")
	if err := florist.CopyFileFromFs(vfs, srcPath, authorizedKeysPath, 0400, owner); err != nil {
		return err
	}

	log.Debug("added file", "path", authorizedKeysPath)
	return nil
}
