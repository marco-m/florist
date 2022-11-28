package consul_test

import (
	"embed"
	"io/fs"
	"testing"

	"github.com/gertd/wild"
	"github.com/google/go-cmp/cmp"
	"github.com/marco-m/florist"
	"github.com/marco-m/florist/flowers/consul"
	"gotest.tools/v3/assert"
)

//go:embed files
var filesFS embed.FS

func TestConsulServerInstallSuccessVM(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)

	files, err := fs.Sub(filesFS, "files")
	assert.NilError(t, err)

	florist.SetLogger(florist.NewLogger("test-consulserver"))
	cs := consul.ServerFlower{
		FilesFS: files,
		Version: "1.11.2",
		Hash:    "380eaff1b18a2b62d8e1d8a7cbc3f3e08b34d3f7187ee335b891ca2ba98784b3",
	}
	assert.NilError(t, cs.Init())

	assert.NilError(t, cs.Install())
}

func TestConsulServerInstallFailureVM(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)

	// FIXME Since not compatible with a client install, I should wipe client installs before...
	// at this point, should I do it as a Flower method, Unistall(), or should I do it grossly only here?

	florist.SetLogger(florist.NewLogger("test-consulserver"))

	testCases := []struct {
		name        string
		flower      consul.ServerFlower
		wantErrWild string // wildcard matching
	}{
		{
			name:        "missing version",
			flower:      consul.ServerFlower{},
			wantErrWild: "* missing version",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.flower.Init()

			if err == nil {
				t.Fatalf("install:\nhave: <no error>\nwant: %s", tc.wantErrWild)
			}

			have := err.Error()
			if !wild.Match(tc.wantErrWild, have, false) {
				diff := cmp.Diff(tc.wantErrWild, have)
				t.Fatalf("error msg wildcard mismatch: (-want +have):\n%s", diff)
			}
		})
	}
}

func TestConsulClientInstallVM(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)

	// FIXME WRITEME
}
