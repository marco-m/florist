package daisy_test

import (
	"io"
	"path/filepath"
	"testing"
	"testing/fstest"
	"time"

	"github.com/marco-m/florist/example/flowers/daisy"
	"github.com/marco-m/florist/pkg/florist"
	"github.com/marco-m/rosina/assert"
)

// TODO: test install all defaults.

func TestDaisyInstall(t *testing.T) {
	// WARNING: In this case we do not call
	// florist.SkipIfNotDisposableHost(t)
	// because this is a special flower!

	err := florist.LowLevelInit(io.Discard, "INFO", time.Hour)
	assert.NoError(t, err, "florist.LowLevelInit")

	fsys := fstest.MapFS{
		daisy.InstallPlainFileSrc: {
			Data: []byte("Johnny Stecchino"),
		},
		daisy.InstallTmplFileSrc: {
			Data: []byte("{{.PetalColor}}"),
		},
	}

	fl := &daisy.Flower{
		Inst: daisy.Inst{Fsys: fsys},
		Conf: daisy.Conf{},
	}
	err = fl.Init()
	assert.NoError(t, err, "fl.Init")

	t.Run("install runs successfully", func(t *testing.T) {
		err = fl.Install()
		assert.NoError(t, err, "fl.Install")
	})

	t.Run("read back what install wrote", func(t *testing.T) {
		// File 1
		{
			fpath := filepath.Join(fl.Inst.DstDir, daisy.InstallPlainFileDst)
			assert.FileEqualsString(t, fpath, "Johnny Stecchino")
		}
		// File 2
		{
			fpath := filepath.Join(fl.Inst.DstDir, daisy.InstallTmplFileDst)
			assert.FileEqualsString(t, fpath, "white")
		}
	})
}

func TestDaisyConfigure(t *testing.T) {
	// WARNING: In this case we do not call
	// florist.SkipIfNotDisposableHost(t)
	// because this is a special flower!

	fsys := fstest.MapFS{
		daisy.ConfigTmplFileSrc: {
			Data: []byte(`{{.PetalColor}} {{.Environment}} {{.GossipKey}}`),
		},
	}

	fl := &daisy.Flower{
		Inst: daisy.Inst{Fsys: fsys},
		Conf: daisy.Conf{
			Environment: "dev",
			GossipKey:   "sesamo",
		},
	}
	err := fl.Init()
	assert.NoError(t, err, "fl.Init")

	t.Run("configure runs successfully", func(t *testing.T) {
		err = fl.Configure()
		assert.NoError(t, err, "fl.Configure")
	})

	t.Run("read back what configure wrote", func(t *testing.T) {
		fpath := filepath.Join(fl.Inst.DstDir, daisy.ConfigTmplFileDst)
		assert.FileEqualsString(t, fpath, "white dev sesamo")
	})
}
