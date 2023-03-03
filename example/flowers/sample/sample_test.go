package sample_test

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/hashicorp/go-hclog"
	"gotest.tools/v3/assert"

	"github.com/marco-m/florist"
	"github.com/marco-m/florist/example/flowers/sample"
	"github.com/marco-m/florist/pkg/installer"
)

func TestSampleInstall(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)

	fsys := fstest.MapFS{
		sample.InstallStaticFileSrc: {
			Data: []byte("Nel mezzo del cammin di nostra vita"),
		},
		sample.InstallTmplFileSrc: {
			Data: []byte("Nel mezzo del {{.Fruit}} di nostra vita"),
		},
	}

	flower := &sample.Flower{}

	t.Run("install runs successfully", func(t *testing.T) {
		log := hclog.NewNullLogger()
		inst := installer.New(log, florist.CacheValidityDefault, fsys)
		assert.NilError(t, inst.AddFlower(flower))

		os.Args = []string{"sut", "install", flower.String()}
		assert.NilError(t, inst.Run())
	})

	t.Run("read back what install wrote", func(t *testing.T) {
		// File 1
		{
			path := filepath.Join(flower.DstDir, sample.InstallStaticFileDst)
			have, err := os.ReadFile(path)
			assert.NilError(t, err)
			assert.Equal(t, string(have), "Nel mezzo del cammin di nostra vita")
		}
		// File 2
		{
			path := filepath.Join(flower.DstDir, sample.InstallTmplFileDst)
			have, err := os.ReadFile(path)
			assert.NilError(t, err)
			assert.Equal(t, string(have), "Nel mezzo del banana di nostra vita")
		}
	})
}

func TestSampleConfigure(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)

	fsys := fstest.MapFS{
		sample.ConfigTmplFileSrc: {
			Data: []byte(`Nel {{index . "embed/secrets/secret"}} del {{.Fruit}} di nostra {{index . "embed/secrets/custom"}}`),
		},
		sample.SecretK: {Data: []byte("mezzo")},
		sample.CustomK: {Data: []byte("vita")},
	}

	flower := &sample.Flower{Fruit: "cammin"}

	t.Run("configure runs successfully", func(t *testing.T) {
		log := hclog.NewNullLogger()
		inst := installer.New(log, florist.CacheValidityDefault, fsys)
		assert.NilError(t, inst.AddFlower(flower))

		os.Args = []string{"sut", "configure", flower.String()}
		assert.NilError(t, inst.Run())
	})

	t.Run("read back what configure wrote", func(t *testing.T) {
		path := filepath.Join(flower.DstDir, sample.ConfigTmplFileDst)
		have, err := os.ReadFile(path)
		assert.NilError(t, err)
		assert.Equal(t, string(have), "Nel mezzo del cammin di nostra vita")
	})
}
