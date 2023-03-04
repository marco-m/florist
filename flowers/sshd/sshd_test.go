package sshd_test

import (
	"embed"
	"io/fs"
	"os"
	"strings"
	"testing"
	"testing/fstest"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"

	"github.com/marco-m/florist"
	"github.com/marco-m/florist/flowers/sshd"
)

//go:embed files
var filesFS embed.FS

func TestSshdInstallSuccess(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)

	florist.SetLogger(florist.NewLogger("test"))

	fl := sshd.Flower{Port: 1234}
	assert.NilError(t, fl.Init(filesFS))

	assert.NilError(t, fl.Install())

	buf, err := os.ReadFile(sshd.ConfigPathDst)
	assert.NilError(t, err)
	have := string(buf)
	want := "Port 1234"
	assert.Assert(t, cmp.Contains(have, want), "reading %s", sshd.ConfigPathDst)
}

func TestSshdConfigureSuccess(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)

	const (
		SshHostEd25519Key        = "-----BEGIN OPENSSH PRIVATE KEY-----\nb3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtzc2gtZW\nQyNTUxOQAAACBSwG6+bwGAOqNANUQiv89cQwVnuzxLLEe5bvVLVqOGVQAAAJDCUHL+wlBy\n/gAAAAtzc2gtZWQyNTUxOQAAACBSwG6+bwGAOqNANUQiv89cQwVnuzxLLEe5bvVLVqOGVQ\nAAAEByQPIGdxQvT+KK9Gb1pZSHOxeKuTFZinuOMqLdUXtTQ1LAbr5vAYA6o0A1RCK/z1xD\nBWe7PEssR7lu9UtWo4ZVAAAACmNvbnRyb2xsZXIBAgM=\n-----END OPENSSH PRIVATE KEY-----\n"
		SshHostEd25519KeyPub     = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIFLAbr5vAYA6o0A1RCK/z1xDBWe7PEssR7lu9UtWo4ZV controller\n"
		SshHostEd25519KeyCertPub = "ssh-ed25519-cert-v01@openssh.com AAAAIHNzaC1lZDI1NTE5LWNlcnQtdjAxQG9wZW5zc2guY29tAAAAICKzG6B7ncoyduo40F9j09SKmNHmN0fBB/88EKhUrKGQAAAAIFLAbr5vAYA6o0A1RCK/z1xDBWe7PEssR7lu9UtWo4ZVAAAAAAAAAAAAAAACAAAAE2NvbnRyb2xsZXItb3Jzb2xhYnMAAAAAAAAAAAAAAAD//////////wAAAAAAAAAAAAAAAAAAADMAAAALc3NoLWVkMjU1MTkAAAAgemiCHSBWFPq5PWhEGrBoOIMAlqNFC/e3kyKsYoYCzyoAAABTAAAAC3NzaC1lZDI1NTE5AAAAQO2pYU1CkGRyQK7PjaE/8r6aoKZEwkLfEtlpoDtmLtfxckMPxh3xPp3K2Jrkkn+2YAi92PYmeHhNEELBd82h6gA= controller\n"
	)

	secretsFS := fstest.MapFS{
		sshd.SshHostEd25519KeySecret:        {Data: []byte(SshHostEd25519Key)},
		sshd.SshHostEd25519KeyPubSecret:     {Data: []byte(SshHostEd25519KeyPub)},
		sshd.SshHostEd25519KeyCertPubSecret: {Data: []byte(SshHostEd25519KeyCertPub)},
	}

	// write a simple unionfs that reads the files from embed and the secrets
	// from the test.mapfs above!
	fsys := &UnionFS{
		FilesFS:   filesFS,
		SecretsFS: secretsFS,
	}

	florist.SetLogger(florist.NewLogger("test"))

	fl := sshd.Flower{Port: 1234}
	assert.NilError(t, fl.Init(fsys))

	assert.NilError(t, fl.Configure())

	// FIXME WRITEME... FOR THE REST OF CONFIGURE (THE KEYS...)
}

// Implements fs.FS. As usual, Go is a pleasure to use.
type UnionFS struct {
	FilesFS   fs.FS
	SecretsFS fs.FS
}

func (ufs *UnionFS) Open(name string) (fs.File, error) {
	// Barely sufficient logic.
	if strings.HasPrefix(name, "files/") {
		return ufs.FilesFS.Open(name)
	}
	return ufs.SecretsFS.Open(name)
}
