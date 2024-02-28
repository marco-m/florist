package ssh

import (
	"fmt"
	"io/fs"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/marco-m/florist/pkg/florist"
)

// AddAuthorizedKeys overwrites HOME/.ssh/authorized_keys with the contents of
// the files matching '*.pub' below fsys, where HOME is the home directory of
// username.
func AddAuthorizedKeys(username string, fsys fs.FS) error {
	log := florist.Log().With("user", username)
	log.Info("adding SSH authorized_keys")

	theUser, err := user.Lookup(username)
	if err != nil {
		return fmt.Errorf("AddAuthorizedKeys: %s", err)
	}

	var keys []string
	matches, err := fs.Glob(fsys, "*.pub")
	if err != nil {
		return fmt.Errorf("AddAuthorizedKeys: %s", err)
	}
	log.Debug("found", "public-keys", matches)
	for _, x := range matches {
		buf, err := fs.ReadFile(fsys, x)
		if err != nil {
			return fmt.Errorf("AddAuthorizedKeys: %s", err)
		}
		keys = append(keys, strings.TrimSpace(string(buf)))
	}

	// If $HOME/.ssh doesn't exist, create it, correct owner and permissions.
	sshDir := filepath.Join(theUser.HomeDir, ".ssh")
	if err := florist.Mkdir(sshDir, username, 0700); err != nil {
		return fmt.Errorf("AddAuthorizedKeys: %s", err)
	}

	authorizedKeysPath := filepath.Join(sshDir, "authorized_keys")
	fi, err := os.OpenFile(authorizedKeysPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("AddAuthorizedKeys: %s", err)
	}
	defer fi.Close()
	uid, _ := strconv.Atoi(theUser.Uid)
	gid, _ := strconv.Atoi(theUser.Gid)
	if err := os.Chown(authorizedKeysPath, uid, gid); err != nil {
		return fmt.Errorf("AddAuthorizedKeys: %s", err)
	}

	if _, err = fi.WriteString(strings.Join(keys, "\n")); err != nil {
		return fmt.Errorf("AddAuthorizedKeys: %s", err)
	}

	log.Debug("added file", "path", authorizedKeysPath, "public-keys", matches)
	return nil
}
