package ssh

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"

	"github.com/marco-m/florist"
)

// This overwrites !
func AddAuthorizedKeys(owner *user.User, contents string) error {
	log := florist.Log.Named("ssh.AddAuthorizedKeys").With("user", owner.Username)

	// If $HOME/.ssh doesn't exist, create it, correct owner and permissions.
	sshDir := filepath.Join(owner.HomeDir, ".ssh")
	if err := florist.Mkdir(sshDir, owner, 0700); err != nil {
		return fmt.Errorf("AddAuthorizedKeys: %s", err)
	}

	authorizedKeysPath := filepath.Join(sshDir, "authorized_keys")
	f, err := os.OpenFile(authorizedKeysPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("AddAuthorizedKeys: %s", err)
	}
	defer f.Close()
	uid, _ := strconv.Atoi(owner.Uid)
	gid, _ := strconv.Atoi(owner.Gid)
	if err := os.Chown(authorizedKeysPath, uid, gid); err != nil {
		return fmt.Errorf("AddAuthorizedKeys: %s", err)
	}

	if _, err = f.WriteString(contents); err != nil {
		return fmt.Errorf("AddAuthorizedKeys: %s", err)
	}

	log.Debug("added file", "path", authorizedKeysPath)
	return nil
}
