package ssh_test

import (
	"embed"
	"testing"

	"github.com/marco-m/florist"
	"github.com/marco-m/florist/pkg/ssh"
)

//go:embed testdata
var filesFS embed.FS

func TestSshAddAuthorizedKeysVM(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)

	user, err := florist.UserAdd("ssh-user")
	if err != nil {
		t.Fatalf("add user:\nhave: %s\nwant: <no error>", err)
	}

	err = ssh.AddAuthorizedKeys(user, filesFS, "testdata/client_key.pub")

	if err != nil {
		t.Fatalf("\nhave: %s\nwant: <no error>", err)
	}
}
