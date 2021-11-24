package florist

import (
	"fmt"
	"os/exec"
	"os/user"
	"strings"
)

// Add user and create home directory.
// Do nothing if user already present.
// If groups not empty, add the user to the supplementary groups.
// Password login is disabled (use SSH public key or use passwd)
func UserAdd(username string, groups ...string) (*user.User, error) {
	log := Log.Named("UserAdd").With("user", username)

	if userinfo, err := user.Lookup(username); err == nil {
		log.Debug("user already present")
		return userinfo, nil
	}
	// Here we should check for the specific user.UnknownUserError
	// but we err on optimist and keep going; in any case adduser will fail
	// if something is wrong...

	cmd := exec.Command(
		"adduser",
		"--gecos", fmt.Sprintf("'user %s'", username),
		"--disabled-password",
		username)
	if err := LogRun(log, cmd); err != nil {
		return nil, fmt.Errorf("user: add: %s", err)
	}
	log.Debug("user added")

	if err := SupplementaryGroups(username, groups...); err != nil {
		return nil, err
	}

	userinfo, err := user.Lookup(username)
	if err != nil {
		return nil, fmt.Errorf("user: lookup: %s", err)
	}

	return userinfo, nil
}

// if username is present, then add it to the supplementary groups
func SupplementaryGroups(username string, groups ...string) error {
	log := Log.Named("SupplementaryGroups").With("user", username).With("groups", groups)
	if _, err := user.Lookup(username); err != nil {
		log.Debug("user not found, skip")
		return nil
	}
	if len(groups) == 0 {
		log.Debug("no supplementary groups, skip")
		return nil
	}
	args := strings.Join(groups, ",")
	cmd := exec.Command("usermod", "--append", "--groups", args, username)
	if err := LogRun(log, cmd); err != nil {
		return fmt.Errorf("user: add to groups: %s", err)
	}
	log.Debug("user added to supplementary groups")

	return nil
}

func UserSystemAdd(username string, homedir string) (*user.User, error) {
	log := Log.Named("UserSystemAdd").With("user", username)

	if userinfo, err := user.Lookup(username); err == nil {
		log.Debug("user already present")
		return userinfo, nil
	}

	cmd := exec.Command(
		"adduser",
		"--system", "--group",
		"--home", homedir,
		"--gecos", fmt.Sprintf("'user %s'", username),
		username)
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

func GroupSystemAdd(groupname string) error {
	log := Log.Named("GroupAdd").With("group", groupname)

	cmd := exec.Command("addgroup", "--system", groupname)
	if err := LogRun(log, cmd); err != nil {
		return fmt.Errorf("group: add: %s", err)
	}
	log.Debug("group added")
	return nil
}
