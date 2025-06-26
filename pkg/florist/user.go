package florist

import (
	"fmt"
	"log/slog"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
)

// UserAddOpts contains the options to be passed to [UserAdd].
type UserAddOpts struct {
	// Create a system account. Default: create a user account.
	System bool
	// The home directory. Default: /home/<username>
	HomeDir string
	// A list of supplementary groups to which the user will be added.
	// The default is for the user to belong only to the initial group.
	Groups []string
	// The UID. Default: choosen by the system. Use [Ptr] to fill it.
	UID *int
}

// UserAdd adds user 'username' and creates its home directory.
// It is not an error if the user is already present.
// Password login is disabled (use SSH public key or use passwd to set).
func UserAdd(username string, opts UserAddOpts) error {
	log := slog.With("user", username)

	log.Info("user-add")
	if _, err := user.Lookup(username); err == nil {
		log.Debug("user-add", "status", "user-already-present")
		return nil
	}
	// Here we should check for the specific error user.UnknownUserError
	// but we err on optimist and keep going; in any case adduser will fail
	// if something is wrong...

	args := []string{
		username,
		fmt.Sprintf("--comment='user %s'", username),
		"--create-home",
	}
	// Options:
	if opts.System {
		args = append(args, "--system")
	}
	if opts.HomeDir != "" {
		args = append(args, "--home-dir="+opts.HomeDir)
	}
	if len(opts.Groups) > 0 {
		args = append(args, "--groups="+strings.Join(opts.Groups, ","))
	}
	if opts.UID != nil {
		args = append(args, "--uid="+strconv.Itoa(*opts.UID))
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

// UserModOpts contains the options to be passed to [UserMod].
type UserModOpts struct {
	// A list of supplementary groups to which the user will be added.
	Groups []string
}

// UserMod modifies 'username' according to 'opts'.
func UserMod(username string, opts UserModOpts) error {
	log := slog.With("user", username)

	args := []string{username}
	// Options:
	if len(opts.Groups) > 0 {
		args = append(args, "--groups="+strings.Join(opts.Groups, ","))
	}

	cmd := exec.Command("usermod", args...)
	if err := CmdRun(log, cmd); err != nil {
		return fmt.Errorf("usermod: %s", err)
	}
	log.Debug("user-mod", "status", "user-modified")

	return nil
}

// GroupAddOpts contains the options to be passed to [GroupAdd].
type GroupAddOpts struct {
	// Create a system group. Default: create a user group.
	System bool
}

// GroupAdd adds group 'groupname'.
// It is not an error if 'groupname' already exists.
func GroupAdd(groupname string, opts GroupAddOpts) error {
	log := slog.With("group", groupname)
	log.Info("group-add")

	args := []string{groupname}
	// Options:
	if opts.System {
		args = append(args, "--system")
	}

	cmd := exec.Command("groupadd", args...)
	if err := CmdRun(log, cmd); err != nil {
		return fmt.Errorf("group: add: %s", err)
	}
	log.Debug("group-added")
	return nil
}
