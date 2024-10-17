package florist

import (
	"fmt"
	"os/exec"
	"os/user"
	"strings"
)

// Add user and create home directory.
// Do nothing if user already present.
// Password login is disabled (use SSH public key or use passwd)
func UserAdd(username string) error {
	log := Log().With("user", username)

	log.Info("user-add")
	if _, err := user.Lookup(username); err == nil {
		log.Debug("user-add", "status", "user-already-present")
		return nil
	}
	// Here we should check for the specific user.UnknownUserError
	// but we err on optimist and keep going; in any case adduser will fail
	// if something is wrong...

	cmd := exec.Command(
		"adduser",
		"--gecos", fmt.Sprintf(`"user %s"`, username),
		"--disabled-password",
		username)
	if err := CmdRun(log, cmd); err != nil {
		return fmt.Errorf("user: add: %s", err)
	}
	log.Debug("user-add", "status", "user-added")

	_, err := user.Lookup(username)
	if err != nil {
		return fmt.Errorf("user: lookup: %s", err)
	}

	return nil
}

// if username is present, then add it to the supplementary groups
func SupplementaryGroups(username string, groups ...string) error {
	log := Log().With("user", username).With("groups", groups)
	if _, err := user.Lookup(username); err != nil {
		log.Debug("supplementary-groups", "status", "user-not-found-skipping")
		return nil
	}
	if len(groups) == 0 {
		log.Debug("supplementary-groups", "status", "no-supplementary-groups-skipping")
		return nil
	}
	args := strings.Join(groups, ",")
	cmd := exec.Command("usermod", "--append", "--groups", args, username)
	if err := CmdRun(log, cmd); err != nil {
		return fmt.Errorf("user: add to groups: %s", err)
	}
	log.Debug("supplementary-groups", "status", "user-added-to-groups")

	return nil
}

func UserSystemAdd(username string, homedir string) error {
	log := Log().With("user", username)

	if _, err := user.Lookup(username); err == nil {
		log.Debug("user-system-add", "status", "user already present")
		return nil
	}

	cmd := exec.Command(
		"adduser",
		"--system", "--group",
		"--home", homedir,
		"--gecos", fmt.Sprintf(`"user %s"`, username),
		username)
	if err := CmdRun(log, cmd); err != nil {
		return fmt.Errorf("user: add: %s", err)
	}
	log.Debug("user-system-add", "status", "user added")

	_, err := user.Lookup(username)
	if err != nil {
		return fmt.Errorf("user: lookup: %s", err)
	}

	return nil
}

func GroupSystemAdd(groupname string) error {
	log := Log().With("group", groupname)

	cmd := exec.Command("addgroup", "--system", groupname)
	if err := CmdRun(log, cmd); err != nil {
		return fmt.Errorf("group: add: %s", err)
	}
	log.Debug("group-added")
	return nil
}
