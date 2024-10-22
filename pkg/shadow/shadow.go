// Package shadow edits the /etc/passwd and /etc/shadow user accounts
// files on Unix systems.
package shadow

import "time"

const (
	Passwd = "/etc/passwd"
	Shadow = "/etc/shadow"
)

type Mode int

const (
	ReadOnly Mode = iota
	ReadWrite
)

func (mo Mode) String() string {
	switch mo {
	case ReadOnly:
		return "ReadOnly"
	case ReadWrite:
		return "ReadWrite"
	default:
		return "Invalid"
	}
}

type PasswdEntry struct {
	LoginName string
	// If the password is “x”, then the encrypted password is stored in the
	// shadow file; if the shadow file doesn't have a corresponding entry, the
	// user account is invalid.
	Password string // encrypted
	UID      int
	GID      int
	Comment  string
	Home     string
	Command  string
}

type ShadowEntry struct {
	LoginName         string
	Password          string // encrypted
	PassLastChanged   time.Time
	MinPassAge        time.Duration
	MaxPassAge        time.Duration
	PassWarning       time.Duration
	PassInactivity    time.Duration
	AccountExpiration time.Time
	Reserved          any
}

type DB struct {
}

func Open(mode Mode) (*DB, error) {
	return OpenFiles(Passwd, Shadow, mode)
}

func OpenFiles(passwd string, shadow string, mode Mode) (*DB, error) {
	// passwd entry:
	// bin:x:2:2:bin:/bin:/usr/sbin/nologin

	// shadow entry:
	// bin:*:19702:0:99999:7:::
}
