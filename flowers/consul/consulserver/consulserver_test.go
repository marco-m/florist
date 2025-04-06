package consulserver_test

import (
	"io"
	"testing"
	"time"

	"github.com/marco-m/florist/flowers/consul/consulserver"
	"github.com/marco-m/florist/pkg/florist"
	"github.com/marco-m/rosina/assert"
)

func TestConsulServerInstallSuccessVM(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)
	err := florist.LowLevelInit(io.Discard, "INFO", time.Hour)
	assert.NoError(t, err, "florist.LowLevelInit")

	fl := consulserver.Flower{
		Inst: consulserver.Inst{
			Version: "1.11.2",
			Hash:    "380eaff1b18a2b62d8e1d8a7cbc3f3e08b34d3f7187ee335b891ca2ba98784b3",
		},
	}
	err = fl.Init()
	assert.NoError(t, err, "fl.Init")

	err = fl.Install()
	assert.NoError(t, err, "fl.Install")
}

func TestConsulServerInstallFailureVM(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)
	err := florist.LowLevelInit(io.Discard, "INFO", time.Hour)
	assert.NoError(t, err, "florist.LowLevelInit")

	// FIXME Since not compatible with a client install, I should wipe client installs before...
	// at this point, should I do it as a Flower method, Unistall(), or should I do it grossly only here?

	type testCase struct {
		name    string
		flower  consulserver.Flower
		wantErr string
	}

	testCases := []testCase{
		{
			name:    "missing version",
			flower:  consulserver.Flower{},
			wantErr: ` missing version`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.flower.Init()
			assert.ErrorContains(t, err, tc.wantErr)
		})
	}
}

func TestConsulClientConfigureVM(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)

	// FIXME WRITEME
	// cfgFile := path.Join(consul.CfgDir, filepath.Base(consulserver.ConfigFile))
	//	assert.FileContains(t,cfgFile, "Port 1234\n")
}
