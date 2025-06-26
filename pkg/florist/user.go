package florist

import (
	"fmt"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
)

// Ptr returns a pointer to 'p'.
// Needed to fill a struct field when the Go zero value cannot be used.
func Ptr[T any](p T) *T {
	return &p
}

// UserAddOpt contains the options to be passed to [UserAdd].
type UserAddOpt struct {
	// Create a system account. Default: create a user account.
	System bool
	// The home directory. Default: /home/<username>
	HomeDir string
	// A list of supplementary groups to which the user will be added.
	// The default is for the user to belong only to the initial group.
	SupplementaryGroups []string
	// The UID. Default: choosen by the system. Use [Ptr] to fill it.
	UID *int
}

// UserAdd adds user 'username' and creates its home directory.
// It is not an error if the user already present.
// Password login is disabled (use SSH public key or use passwd).
func UserAdd(username string, opt UserAddOpt) error {
	log := Log().With("user", username)

	log.Info("user-add")
	if _, err := user.Lookup(username); err == nil {
		log.Debug("user-add", "status", "user-already-present")
		return nil
	}
	// Here we should check for the specific user.UnknownUserError
	// but we err on optimist and keep going; in any case adduser will fail
	// if something is wrong...

	args := []string{
		username,
		"--comment", fmt.Sprintf(`"user %s"`, username),
	}
	// Options:
	if opt.System {
		args = append(args, "--system")
	}
	if opt.HomeDir != "" {
		args = append(args, "--home-dir="+opt.HomeDir)
	}
	args = append(args, "--create-home")
	if len(opt.SupplementaryGroups) > 0 {
		args = append(args, "--groups="+strings.Join(opt.SupplementaryGroups, ","))
	}
	if opt.UID != nil {
		args = append(args, "--uid="+strconv.Itoa(*opt.UID))
	}

	cmd := exec.Command("useradd", args...)
	if err := CmdRun(log, cmd); err != nil {
		return fmt.Errorf("UserAdd: %s (%s)", err, cmd)
	}
	log.Debug("user-add", "status", "user-added")

	// Should be impossible to fail, but you never know.
	_, err := user.Lookup(username)
	if err != nil {
		return fmt.Errorf("user: lookup: %s", err)
	}

	return nil
}

// UserModOpt contains the options to be passed to [UserMod].
type UserModOpt struct {
	// A list of supplementary groups to which the user will be added.
	SupplementaryGroups []string
}

// UserMod modifies 'username' according to 'opts'.
func UserMod(username string, opt UserModOpt) error {
	log := Log().With("user", username)

	args := []string{username}
	// Options:
	if len(opt.SupplementaryGroups) > 0 {
		args = append(args, "--groups="+strings.Join(opt.SupplementaryGroups, ","))
	}

	cmd := exec.Command("usermod", args...)
	if err := CmdRun(log, cmd); err != nil {
		return fmt.Errorf("usermod: %s", err)
	}
	log.Debug("user-mod", "status", "user-modified")

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
