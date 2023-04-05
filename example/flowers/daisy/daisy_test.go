package daisy_test

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/hashicorp/go-hclog"
	"gotest.tools/v3/assert"

	"github.com/marco-m/florist/example/flowers/daisy"

	"github.com/marco-m/florist/pkg/florist"
	"github.com/marco-m/florist/pkg/provisioner"
)

func TestSampleInstall(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)

	flower := &daisy.Flower{}

	fsys := fstest.MapFS{
		path.Join(flower.String(), daisy.InstallStaticFileSrc): {
			Data: []byte("Nel mezzo del cammin di nostra vita"),
		},
		path.Join(flower.String(), daisy.InstallTplFileSrc): {
			Data: []byte("Nel mezzo del {{.Fruit}} di nostra vita"),
		},
	}

	t.Run("install runs successfully", func(t *testing.T) {
		log := hclog.NewNullLogger()
		inst, err := provisioner.New(log, florist.CacheValidity, fsys, fsys)
		assert.NilError(t, err)
		assert.NilError(t, inst.AddBouquet("name", "desc", flower))

		os.Args = []string{"sut", "install", "name"}
		err = inst.Run()
		assert.NilError(t, err, "\n%s", listFiles(fsys, fsys))
	})

	t.Run("read back what install wrote", func(t *testing.T) {
		// File 1
		{
			path := filepath.Join(flower.DstDir, daisy.InstallStaticFileDst)
			have, err := os.ReadFile(path)
			assert.NilError(t, err)
			assert.Equal(t, string(have), "Nel mezzo del cammin di nostra vita")
		}
		// File 2
		{
			path := filepath.Join(flower.DstDir, daisy.InstallTmplFileDst)
			have, err := os.ReadFile(path)
			assert.NilError(t, err)
			assert.Equal(t, string(have), "Nel mezzo del banana di nostra vita")
		}
	})
}

func listFiles(files, secrets fs.FS) string {
	s, _ := provisioner.ListFs(files, secrets)
	return s
}

func TestSampleConfigure(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)

	flower := &daisy.Flower{Fruit: "cammin"}

	files := fstest.MapFS{
		path.Join(flower.String(), daisy.ConfigTplFileSrc): {
			Data: []byte(`Nel {{.Secret}} del {{.Fruit}} di nostra {{.Custom}}`),
		},
	}
	secrets := fstest.MapFS{
		path.Join("flowers", flower.String(), "secret"): {
			Data: []byte("mezzo"),
		},
		path.Join("flowers", flower.String(), "custom"): {
			Data: []byte("vita"),
		},
	}

	t.Run("configure runs successfully", func(t *testing.T) {
		log := hclog.NewNullLogger()
		inst, err := provisioner.New(log, florist.CacheValidity, files, secrets)
		assert.NilError(t, err)
		assert.NilError(t, inst.AddBouquet("name", "desc", flower))

		os.Args = []string{"sut", "configure", "name"}
		err = inst.Run()
		assert.NilError(t, err, "\n%s", listFiles(files, secrets))
	})

	t.Run("read back what configure wrote", func(t *testing.T) {
		path := filepath.Join(flower.DstDir, daisy.ConfigTplFileDst)
		have, err := os.ReadFile(path)
		assert.NilError(t, err)
		assert.Equal(t, string(have), "Nel mezzo del cammin di nostra vita")
	})
}
