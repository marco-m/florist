package sshd_test

import (
	"io"
	"testing"
	"time"

	"github.com/go-quicktest/qt"

	"github.com/marco-m/rosina"

	"github.com/marco-m/florist/flowers/sshd"
	"github.com/marco-m/florist/pkg/florist"
)

func TestSshdInstallSuccess(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)

	err := florist.LowLevelInit(io.Discard, "INFO", time.Hour)
	qt.Assert(t, qt.IsNil(err))

	fl := sshd.Flower{
		Inst: sshd.Inst{},
	}
	err = fl.Init()
	qt.Assert(t, qt.IsNil(err))

	err = fl.Install()
	qt.Assert(t, qt.IsNil(err))

	qt.Assert(t, rosina.FileContains(sshd.SshdConfigDst, "Port 22\n"))
}

func TestSshdConfigureSuccess(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)

	const (
		SshHostEd25519Key        = "-----BEGIN OPENSSH PRIVATE KEY-----\nb3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtzc2gtZW\nQyNTUxOQAAACBSwG6+bwGAOqNANUQiv89cQwVnuzxLLEe5bvVLVqOGVQAAAJDCUHL+wlBy\n/gAAAAtzc2gtZWQyNTUxOQAAACBSwG6+bwGAOqNANUQiv89cQwVnuzxLLEe5bvVLVqOGVQ\nAAAEByQPIGdxQvT+KK9Gb1pZSHOxeKuTFZinuOMqLdUXtTQ1LAbr5vAYA6o0A1RCK/z1xD\nBWe7PEssR7lu9UtWo4ZVAAAACmNvbnRyb2xsZXIBAgM=\n-----END OPENSSH PRIVATE KEY-----\n"
		SshHostEd25519KeyPub     = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIFLAbr5vAYA6o0A1RCK/z1xDBWe7PEssR7lu9UtWo4ZV controller\n"
		SshHostEd25519KeyCertPub = "ssh-ed25519-cert-v01@openssh.com AAAAIHNzaC1lZDI1NTE5LWNlcnQtdjAxQG9wZW5zc2guY29tAAAAICKzG6B7ncoyduo40F9j09SKmNHmN0fBB/88EKhUrKGQAAAAIFLAbr5vAYA6o0A1RCK/z1xDBWe7PEssR7lu9UtWo4ZVAAAAAAAAAAAAAAACAAAAE2NvbnRyb2xsZXItb3Jzb2xhYnMAAAAAAAAAAAAAAAD//////////wAAAAAAAAAAAAAAAAAAADMAAAALc3NoLWVkMjU1MTkAAAAgemiCHSBWFPq5PWhEGrBoOIMAlqNFC/e3kyKsYoYCzyoAAABTAAAAC3NzaC1lZDI1NTE5AAAAQO2pYU1CkGRyQK7PjaE/8r6aoKZEwkLfEtlpoDtmLtfxckMPxh3xPp3K2Jrkkn+2YAi92PYmeHhNEELBd82h6gA= controller\n"
	)
	err := florist.LowLevelInit(io.Discard, "INFO", time.Hour)
	qt.Assert(t, qt.IsNil(err))

	fl := sshd.Flower{
		Inst: sshd.Inst{},
		Conf: sshd.Conf{
			Port:                     1234,
			SshHostEd25519Key:        SshHostEd25519Key,
			SshHostEd25519KeyPub:     SshHostEd25519KeyPub,
			SshHostEd25519KeyCertPub: SshHostEd25519KeyCertPub,
		},
	}
	err = fl.Init()
	qt.Assert(t, qt.IsNil(err))

	err = fl.Configure()
	qt.Assert(t, qt.IsNil(err))

	qt.Assert(t, rosina.FileEqualsString(sshd.SshHostEd25519KeyDst,
		SshHostEd25519Key))
	qt.Assert(t, rosina.FileEqualsString(sshd.SshHostEd25519KeyPubDst,
		SshHostEd25519KeyPub))
	qt.Assert(t, rosina.FileEqualsString(sshd.SshHostEd25519KeyCertPubDst,
		SshHostEd25519KeyCertPub))
}
