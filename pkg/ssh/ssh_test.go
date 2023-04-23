package ssh_test

import (
	"testing"

	"gotest.tools/v3/assert"

	"github.com/marco-m/florist/pkg/florist"
	"github.com/marco-m/florist/pkg/ssh"
)

func TestSshAddAuthorizedKeysVM(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)
	username := "ssh-user"

	err := florist.UserAdd(username)
	assert.NilError(t, err)

	content := "hello"
	err = ssh.AddAuthorizedKeys(username, content)
	assert.NilError(t, err)
}
