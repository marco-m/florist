package sshd_test

import (
	"embed"
	"io/fs"
	"os"
	"testing"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"

	"github.com/marco-m/florist"
	"github.com/marco-m/florist/flowers/sshd"
)

//go:embed files
var fsys embed.FS

func TestSshdInstallSuccess(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)

	florist.SetLogger(florist.NewLogger("test"))
	fsys, err := fs.Sub(fsys, "files")
	assert.NilError(t, err)

	fl := sshd.Flower{Port: 1234}
	assert.NilError(t, fl.Init(fsys))

	assert.NilError(t, fl.Install())

	buf, err := os.ReadFile(fl.DstSshdConfigPath)
	have := string(buf)
	want := "Port 1234"
	assert.NilError(t, err)
	assert.Assert(t, cmp.Contains(have, want), "reading %s", fl.DstSshdConfigPath)
}

func TestSshdConfigureSuccess(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)

	florist.SetLogger(florist.NewLogger("test"))
	files, err := fs.Sub(fsys, "files")
	assert.NilError(t, err)

	fl := sshd.Flower{
		Port:                     1234,
		SshHostEd25519KeyPub:     "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIFLAbr5vAYA6o0A1RCK/z1xDBWe7PEssR7lu9UtWo4ZV controller\n",
		SshHostEd25519Key:        "-----BEGIN OPENSSH PRIVATE KEY-----\nb3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtzc2gtZW\nQyNTUxOQAAACBSwG6+bwGAOqNANUQiv89cQwVnuzxLLEe5bvVLVqOGVQAAAJDCUHL+wlBy\n/gAAAAtzc2gtZWQyNTUxOQAAACBSwG6+bwGAOqNANUQiv89cQwVnuzxLLEe5bvVLVqOGVQ\nAAAEByQPIGdxQvT+KK9Gb1pZSHOxeKuTFZinuOMqLdUXtTQ1LAbr5vAYA6o0A1RCK/z1xD\nBWe7PEssR7lu9UtWo4ZVAAAACmNvbnRyb2xsZXIBAgM=\n-----END OPENSSH PRIVATE KEY-----\n",
		SshHostEd25519KeyCertPub: "ssh-ed25519-cert-v01@openssh.com AAAAIHNzaC1lZDI1NTE5LWNlcnQtdjAxQG9wZW5zc2guY29tAAAAICKzG6B7ncoyduo40F9j09SKmNHmN0fBB/88EKhUrKGQAAAAIFLAbr5vAYA6o0A1RCK/z1xDBWe7PEssR7lu9UtWo4ZVAAAAAAAAAAAAAAACAAAAE2NvbnRyb2xsZXItb3Jzb2xhYnMAAAAAAAAAAAAAAAD//////////wAAAAAAAAAAAAAAAAAAADMAAAALc3NoLWVkMjU1MTkAAAAgemiCHSBWFPq5PWhEGrBoOIMAlqNFC/e3kyKsYoYCzyoAAABTAAAAC3NzaC1lZDI1NTE5AAAAQO2pYU1CkGRyQK7PjaE/8r6aoKZEwkLfEtlpoDtmLtfxckMPxh3xPp3K2Jrkkn+2YAi92PYmeHhNEELBd82h6gA= controller\n",
	}
	assert.NilError(t, fl.Init(files))

	assert.NilError(t, fl.Configure())

	// FIXME WRITEME... FOR THE REST OF CONFIGURE (THE KEYS...)
}
