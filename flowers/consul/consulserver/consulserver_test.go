package consulserver_test

import (
	"testing"

	"github.com/marco-m/florist/flowers/consul/consulserver"
	"github.com/marco-m/florist/pkg/testhelpers"
	"github.com/marco-m/florist/pkg/florist"
	"github.com/marco-m/rosina/assert"
)

func TestConsulServerInstallSuccessVM(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)
	tempDir := t.TempDir()
	gdn := testhelpers.InitFlorist(t, tempDir)

	fl := consulserver.Flower{
		Version: "2.0.4",
		Hash:    "7a28033850a24fd411722593931625d8b548a27646c3ab70c1379ea7fd2af423",
	}
	err := fl.Init(gdn)
	assert.NoError(t, err, "fl.Init")

	err = fl.Install()
	assert.NoError(t, err, "fl.Install")
}

func TestConsulServerInstallFailureVM(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)
	tempDir := t.TempDir()
	gdn := testhelpers.InitFlorist(t, tempDir)

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
			err := tc.flower.Init(gdn)
			assert.ErrorContains(t, err, tc.wantErr, "Flower.Init")
		})
	}
}

func TestConsulClientConfigureVM(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)

	// FIXME WRITEME
	// cfgFile := path.Join(consul.CfgDir, filepath.Base(consulserver.ConfigFile))
	//	assert.FileContains(t,cfgFile, "Port 1234\n")
}
