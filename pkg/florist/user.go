package florist

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strconv"
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

// SupplementaryGroups adds 'username' to the supplementary groups 'groups'.
// It is an error if any of 'groups' does not exist (create them beforehand
// with GroupSystemAdd).
func SupplementaryGroups(username string, groups ...string) error {
	log := Log().With("user", username).With("groups", groups)

	if len(groups) == 0 {
		return fmt.Errorf("supplementary-groups: must specify at least one group (user: %s)",
			username)
	}
	if _, err := user.Lookup(username); err != nil {
		return fmt.Errorf("supplementary-groups: user not found (user: %s, groups: %s)",
			username, groups)
	}

	args := strings.Join(groups, ",")
	cmd := exec.Command("usermod", "--append", "--groups", args, username)
	if err := CmdRun(log, cmd); err != nil {
		return fmt.Errorf("user: add to groups: %s (user: %s, groups: %s)",
			err, username, groups)
	}
	log.Debug("supplementary-groups", "status", "user-added-to-groups")

	return nil
}

// UserSystemAdd adds the system user 'username' with home directory 'homedir'.
// The user group will be 'username' and the mode of the home directory will be 0o755.
// See also [UserSystemAddWithUID].
func UserSystemAdd(username string, homedir string) error {
	return UserSystemAddWithUID(username, homedir, -1)
}

// UserSystemAddWithUID adds the system user 'username' with home directory 'homedir'
// and user ID 'uid'. In case the user ID is already taken, UserSystemAddWithUID will fail.
// The user group will be 'username' and the mode of the home directory will be 0o755.
// Normally this is not the right function to use: see [UserSystemAdd] instead.
func UserSystemAddWithUID(username string, homedir string, uid int) error {
	log := Log().With("user", username)

	if _, err := user.Lookup(username); err == nil {
		log.Debug("user-system-add", "status", "user already present")
		return nil
	}

	cmdFlags := []string{
		username,
		"--system",
		"--group",
		"--home", homedir,
		"--gecos", fmt.Sprintf(`"user %s"`, username),
	}
	if uid != -1 {
		cmdFlags = append(cmdFlags, "--uid", strconv.Itoa(uid))
	}

	cmd := exec.Command("adduser", cmdFlags...)
	if err := CmdRun(log, cmd); err != nil {
		return fmt.Errorf("user: add: %s", err)
	}
	log.Debug("user-system-add", "status", "user added")

	newUser, err := user.Lookup(username)
	if err != nil {
		return fmt.Errorf("user: lookup: %s", err)
	}

	if newUser.HomeDir != homedir {
		return fmt.Errorf("user %s: homedir: have: %s, want: %s",
			username, newUser.HomeDir, homedir)
	}
	if err := os.Chmod(newUser.HomeDir, 0o755); err != nil {
		return fmt.Errorf("user %s: %s", username, err)
	}

	return nil
}

// GroupSystemAdd adds group 'groupname'. It is not an error if 'groupname'
// already exists.
func GroupSystemAdd(groupname string) error {
	log := Log().With("group", groupname)

	cmd := exec.Command("addgroup", "--system", groupname)
	if err := CmdRun(log, cmd); err != nil {
		return fmt.Errorf("group: add: %s", err)
	}
	log.Debug("group-added")
	return nil
}
