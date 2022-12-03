package sshd_test

import (
	"embed"
	"io/fs"
	"os"
	"strings"
	"testing"

	"github.com/marco-m/florist"
	"github.com/marco-m/florist/flowers/sshd"
	"gotest.tools/v3/assert"
)

//go:embed files
var filesFS embed.FS

func TestSshdInstallSuccess(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)

	files, err := fs.Sub(filesFS, "files")
	assert.NilError(t, err)

	florist.SetLogger(florist.NewLogger("test-sshd"))
	fl := sshd.Flower{FilesFS: files, Port: 1234}
	assert.NilError(t, fl.Init())

	assert.NilError(t, fl.Install())

	buf, err := os.ReadFile(fl.DstSshdConfig)
	have := string(buf)
	want := "Port 1234"
	assert.NilError(t, err)
	assert.Assert(t, strings.Contains(have, want))
}
