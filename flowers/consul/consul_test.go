package consul_test

import (
	"embed"
	"io/fs"
	"path"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/marco-m/florist/flowers/consul"
	"github.com/marco-m/florist/pkg/florist"
)

//go:embed files
var filesFS embed.FS

func TestConsulServerInstallSuccessVM(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)

	florist.SetLogger(florist.NewLogger("test-consulserver"))
	fl := consul.ServerFlower{
		Version: "1.11.2",
		Hash:    "380eaff1b18a2b62d8e1d8a7cbc3f3e08b34d3f7187ee335b891ca2ba98784b3",
	}
	assert.NilError(t, fl.Init())

	files, err := fs.Sub(filesFS, path.Join("files", fl.String()))
	assert.NilError(t, err)
	finder := florist.NewFsFinder(files)
	assert.NilError(t, fl.Install(files, finder))
}

func TestConsulServerInstallFailureVM(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)

	// FIXME Since not compatible with a client install, I should wipe client installs before...
	// at this point, should I do it as a Flower method, Unistall(), or should I do it grossly only here?

	florist.SetLogger(florist.NewLogger("test-consulserver"))

	testCases := []struct {
		name    string
		flower  consul.ServerFlower
		wantErr string
	}{
		{
			name:    "missing version",
			flower:  consul.ServerFlower{},
			wantErr: " missing version",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.flower.Init()
			assert.ErrorContains(t, err, tc.wantErr)
		})
	}
}

func TestConsulClientInstallVM(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)

	// FIXME WRITEME
}
