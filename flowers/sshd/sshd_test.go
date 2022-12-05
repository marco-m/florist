package sshd_test

import (
	"embed"
	"io/fs"
	"os"
	"testing"

	"github.com/marco-m/florist"
	"github.com/marco-m/florist/flowers/sshd"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
)

//go:embed files
var filesFS embed.FS

func TestSshdInstallSuccess(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)

	florist.SetLogger(florist.NewLogger("test"))
	files, err := fs.Sub(filesFS, "files")
	assert.NilError(t, err)

	fl := sshd.Flower{FilesFS: files, Port: 1234}
	assert.NilError(t, fl.Init())

	assert.NilError(t, fl.Install())

	buf, err := os.ReadFile(fl.DstSshdConfigPath)
	have := string(buf)
	want := "Port 1234"
	assert.NilError(t, err)
	assert.Assert(t, cmp.Contains(have, want), "reading %s", fl.DstSshdConfigPath)
}

// FIXME WRITEME...
func TestSshdConfigureSuccess(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)

	florist.SetLogger(florist.NewLogger("test"))
	files, err := fs.Sub(filesFS, "files")
	assert.NilError(t, err)

	fl := sshd.Flower{
		FilesFS: files,
		Port:    1234,
	}
	assert.NilError(t, fl.Init())

	rawFlower, err := os.ReadFile("testdata/florist.config.json")
	assert.NilError(t, err)
	assert.NilError(t, fl.Configure(rawFlower))

	// FIXME WRITEME... FOR THE REST OF CONFIGURE (THE KEYS...)
}
