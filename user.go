package florist

import (
	"fmt"
	"os/exec"
	"os/user"
)

// Add user and create home directory.
// Password login is disabled (use SSH public key or use passwd)
func UserAdd(username string) (*user.User, error) {
	log := Log.Named("UserAdd").With("user", username)

	if userinfo, err := user.Lookup(username); err == nil {
		log.Debug("user already present")
		return userinfo, nil
	}
	// Here we should check for the specific user.UnknownUserError
	// but we err on optimist and keep going; in any case adduser will fail
	// if something is wrong...

	cmd := exec.Command("adduser", "--disabled-password", username)
	if err := LogRun(log, cmd); err != nil {
		return nil, fmt.Errorf("user: add: %s", err)
	}
	log.Debug("user added")

	userinfo, err := user.Lookup(username)
	if err != nil {
		return nil, fmt.Errorf("user: lookup: %s", err)
	}

	return userinfo, nil
}
