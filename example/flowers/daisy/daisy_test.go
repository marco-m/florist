package daisy_test

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/go-quicktest/qt"

	"github.com/marco-m/florist/example/flowers/daisy"
	"github.com/marco-m/florist/pkg/florist"
)

// TODO: test install all defaults.

func TestDaisyInstall(t *testing.T) {
	// WARNING: In this case we do not call
	// florist.SkipIfNotDisposableHost(t)
	// because this is a special flower!

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
	err := fl.Init()
	qt.Assert(t, qt.IsNil(err))

	t.Run("install runs successfully", func(t *testing.T) {
		err := florist.Init(nil)
		qt.Assert(t, qt.IsNil(err))

		err = fl.Install()
		qt.Assert(t, qt.IsNil(err))
	})

	t.Run("read back what install wrote", func(t *testing.T) {
		// File 1
		{
			fpath := filepath.Join(fl.Inst.DstDir, daisy.InstallPlainFileDst)
			have, err := os.ReadFile(fpath)
			qt.Assert(t, qt.IsNil(err))
			qt.Assert(t, qt.Equals(string(have), "Johnny Stecchino"))
		}
		// File 2
		{
			fpath := filepath.Join(fl.Inst.DstDir, daisy.InstallTmplFileDst)
			have, err := os.ReadFile(fpath)
			qt.Assert(t, qt.IsNil(err))
			qt.Assert(t, qt.Equals(string(have), "white"))
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
	qt.Assert(t, qt.IsNil(err))

	t.Run("configure runs successfully", func(t *testing.T) {
		err := florist.Init(nil)
		qt.Assert(t, qt.IsNil(err))

		err = fl.Configure()
		qt.Assert(t, qt.IsNil(err))
	})

	t.Run("read back what configure wrote", func(t *testing.T) {
		fpath := filepath.Join(fl.Inst.DstDir, daisy.ConfigTmplFileDst)
		have, err := os.ReadFile(fpath)

		qt.Assert(t, qt.IsNil(err))
		qt.Assert(t, qt.Equals(string(have), "white dev sesamo"))
	})
}
