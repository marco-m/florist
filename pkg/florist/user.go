package florist

import (
	"fmt"
	"log/slog"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
)

// UserAddArgs contains the arguments for [UserAdd].
type UserAddArgs struct {
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
func UserAdd(username string, args *UserAddArgs) error {
	log := slog.With("user", username)

	log.Info("user-add")
	if _, err := user.Lookup(username); err == nil {
		log.Debug("user-add", "status", "user-already-present")
		return nil
	}
	// Here we should check for the specific error user.UnknownUserError
	// but we err on optimist and keep going; in any case adduser will fail
	// if something is wrong...

	cmdline := []string{
		username,
		fmt.Sprintf("--comment='user %s'", username),
		"--create-home",
	}
	// Arguments.
	if args == nil {
		args = &UserAddArgs{}
	}
	if args.System {
		cmdline = append(cmdline, "--system")
	}
	if args.HomeDir != "" {
		cmdline = append(cmdline, "--home-dir="+args.HomeDir)
	}
	if len(args.Groups) > 0 {
		cmdline = append(cmdline, "--groups="+strings.Join(args.Groups, ","))
	}
	if args.UID != nil {
		cmdline = append(cmdline, "--uid="+strconv.Itoa(*args.UID))
	}

	cmd := exec.Command("useradd", cmdline...)
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

// UserModArgs contains the arguments for [UserMod].
type UserModArgs struct {
	// A list of supplementary groups to which the user will be added.
	Groups []string
}

// UserMod modifies 'username' according to 'args'.
func UserMod(username string, args *UserModArgs) error {
	log := slog.With("user", username)

	cmdline := []string{username}
	// Arguments.
	if args == nil {
		args = &UserModArgs{}
	}
	if len(args.Groups) > 0 {
		cmdline = append(cmdline, "--groups="+strings.Join(args.Groups, ","))
	}

	cmd := exec.Command("usermod", cmdline...)
	if err := CmdRun(log, cmd); err != nil {
		return fmt.Errorf("usermod: %s", err)
	}
	log.Debug("user-mod", "status", "user-modified")

	return nil
}

// GroupAddArgs contains the arguments for [GroupAdd].
type GroupAddArgs struct {
	// Create a system group. Default: create a user group.
	System bool
}

// GroupAdd adds group 'groupname'.
// It is not an error if 'groupname' already exists.
func GroupAdd(groupname string, args *GroupAddArgs) error {
	log := slog.With("group", groupname)
	log.Info("group-add")

	cmdline := []string{groupname}
	// Arguments.
	if args == nil {
		args = &GroupAddArgs{}
	}
	if args.System {
		cmdline = append(cmdline, "--system")
	}

	cmd := exec.Command("groupadd", cmdline...)
	if err := CmdRun(log, cmd); err != nil {
		return fmt.Errorf("group: add: %s", err)
	}
	log.Debug("group-added")
	return nil
}
