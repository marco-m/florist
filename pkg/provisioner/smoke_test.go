package provisioner_test

import (
	"io/fs"
	"os"
	"testing"
	"testing/fstest"

	"github.com/hashicorp/go-hclog"
	"github.com/rogpeppe/go-internal/testscript"

	"github.com/marco-m/florist/pkg/florist"
	"github.com/marco-m/florist/pkg/provisioner"
)

type smokeFlower struct {
	kv map[string]string
}

func (fl *smokeFlower) String() string {
	return "smoke"
}

func (fl *smokeFlower) Description() string {
	return "I am a smoke flower"
}

func (fl *smokeFlower) Init() error {
	fl.kv = make(map[string]string)
	return nil
}

func (fl *smokeFlower) Install(files fs.FS, finder florist.Finder) error {
	return nil
}

func (fl *smokeFlower) Configure(files fs.FS, finder florist.Finder) error {
	return nil
}

func prepare1(log hclog.Logger) (*provisioner.Provisioner, error) {
	files := fstest.MapFS{"smoke/smoke1": {Data: []byte("A")}}
	secrets := fstest.MapFS{
		"base/base1":            {Data: []byte("B")},
		"flowers/smoke/flower1": {Data: []byte("C")},
		"nodes/x/smoke/node1":   {Data: []byte("D")},
	}

	prov, err := provisioner.New(log, florist.CacheValidity, files, secrets)
	if err != nil {
		return nil, err
	}
	prov.UseWorkdir()
	fl := &smokeFlower{}
	err = prov.AddBouquet("x", "stuff for node x", fl)
	if err != nil {
		return nil, err
	}

	return prov, nil
}

func prepare2(log hclog.Logger) (*provisioner.Provisioner, error) {
	fsys := fstest.MapFS{}

	prov, err := provisioner.New(log, florist.CacheValidity, fsys, fsys)
	if err != nil {
		return nil, err
	}
	prov.UseWorkdir()
	fl := &smokeFlower{}
	err = prov.AddBouquet("y", "stuff for node y", fl)
	if err != nil {
		return nil, err
	}

	return prov, nil
}

func TestScriptProvisioner(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "testdata",
	})
}

func TestMain(m *testing.M) {
	log := florist.NewLogger("test")
	os.Exit(testscript.RunMain(m, map[string]func() int{
		"provisioner1": func() int { return provisioner.Main(log, prepare1) },
		"provisioner2": func() int { return provisioner.Main(log, prepare2) },
	}))
}
