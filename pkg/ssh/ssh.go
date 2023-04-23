package ssh

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"

	"github.com/marco-m/florist/pkg/florist"
)

// This overwrites !
func AddAuthorizedKeys(owner string, contents string) error {
	log := florist.Log.Named("ssh.AddAuthorizedKeys").With("user", owner)

	ownerUser, err := user.Lookup(owner)
	if err != nil {
		return fmt.Errorf("florist.copyfile: %s", err)
	}

	// If $HOME/.ssh doesn't exist, create it, correct owner and permissions.
	sshDir := filepath.Join(ownerUser.HomeDir, ".ssh")
	if err := florist.Mkdir(sshDir, owner, 0700); err != nil {
		return fmt.Errorf("AddAuthorizedKeys: %s", err)
	}

	authorizedKeysPath := filepath.Join(sshDir, "authorized_keys")
	f, err := os.OpenFile(authorizedKeysPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("AddAuthorizedKeys: %s", err)
	}
	defer f.Close()
	uid, _ := strconv.Atoi(ownerUser.Uid)
	gid, _ := strconv.Atoi(ownerUser.Gid)
	if err := os.Chown(authorizedKeysPath, uid, gid); err != nil {
		return fmt.Errorf("AddAuthorizedKeys: %s", err)
	}

	if _, err = f.WriteString(contents); err != nil {
		return fmt.Errorf("AddAuthorizedKeys: %s", err)
	}

	log.Debug("added file", "path", authorizedKeysPath)
	return nil
}
