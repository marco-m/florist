package provisioner_test

import (
	"fmt"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/hashicorp/go-hclog"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"

	"github.com/marco-m/florist/pkg/florist"
	"github.com/marco-m/florist/pkg/provisioner"
)

type mockFlower struct {
	Name string
}

func (fl *mockFlower) String() string {
	return fl.Name
}

func (fl *mockFlower) Description() string {
	return "I am a mock flower"
}

func (fl *mockFlower) Init() error {
	return nil
}

func (fl *mockFlower) Install(files fs.FS, finder florist.Finder) error {
	return nil
}

func (fl *mockFlower) Configure(files fs.FS, finder florist.Finder) error {
	return nil
}

func TestProvisionerAddBouquetSuccess(t *testing.T) {
	log := hclog.NewNullLogger()
	fsys := fstest.MapFS{}
	inst, err := provisioner.New(log, florist.CacheValidity, fsys, fsys)
	assert.NilError(t, err)

	flowers := []florist.Flower{
		&mockFlower{Name: "a"},
		&mockFlower{Name: "b"},
		&mockFlower{Name: "c"},
	}

	err = inst.AddBouquet("goofy", "mickey's friend", flowers...)
	assert.NilError(t, err)

	have := inst.Bouquets()
	want := []provisioner.Bouquet{
		{
			Name:        "goofy",
			Description: "mickey's friend",
			Flowers:     flowers,
		},
	}
	assert.Assert(t, cmp.DeepEqual(have, want))
}

func TestProvisionerAddBouquetFailure(t *testing.T) {
	log := hclog.NewNullLogger()
	fsys := fstest.MapFS{}

	flowers := []florist.Flower{
		&mockFlower{Name: "a"},
		&mockFlower{Name: "b"},
		&mockFlower{Name: "c"},
	}

	testCases := []struct {
		name         string
		bouquet      []florist.Flower
		bname        string
		bdescription string
		wantErr      string
	}{
		{
			name:    "name is empty",
			bouquet: flowers,
			wantErr: "AddBouquet: name cannot be empty",
		},
		{
			name:    "desc is empty",
			bouquet: flowers,
			bname:   "foo",
			wantErr: "AddBouquet foo: description cannot be empty",
		},
		{
			name: "flower with empty name",
			bouquet: []florist.Flower{
				&mockFlower{Name: "a"},
				&mockFlower{},
			},
			bname:        "foo",
			bdescription: "bar",
			wantErr:      "AddBouquet foo: flower at position 1 has empty name",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			inst, err := provisioner.New(log, florist.CacheValidity, fsys, fsys)
			assert.NilError(t, err)

			err = inst.AddBouquet(tc.bname, tc.bdescription, tc.bouquet...)

			assert.ErrorContains(t, err, tc.wantErr)
		})
	}
}

func TestProvisionerDuplicateBouquetName(t *testing.T) {
	log := hclog.NewNullLogger()
	fsys := fstest.MapFS{}
	inst, err := provisioner.New(log, florist.CacheValidity, fsys, fsys)
	assert.NilError(t, err)

	bname := "pippo"
	bouquet1 := []florist.Flower{&mockFlower{Name: "1"}}
	bouquet2 := []florist.Flower{&mockFlower{Name: "2"}}
	wantErr := "AddBouquet: there is already a bouquet with name pippo"

	err = inst.AddBouquet(bname, "topolino", bouquet1...)
	assert.NilError(t, err)

	err = inst.AddBouquet(bname, "clarabella", bouquet2...)
	assert.ErrorContains(t, err, wantErr)
}

type spyFlower struct {
	kv map[string]string
}

func (fl *spyFlower) String() string {
	return "spy"
}

func (fl *spyFlower) Description() string {
	return "I am a spy flower"
}

func (fl *spyFlower) Init() error {
	fl.kv = make(map[string]string)
	return nil
}

func (fl *spyFlower) Install(files fs.FS, finder florist.Finder) error {
	return nil
}

func (fl *spyFlower) Configure(files fs.FS, finder florist.Finder) error {
	keys, err := finder.Keys()
	if err != nil {
		return fmt.Errorf("spy.Configure: %s", err)
	}
	for _, k := range keys {
		fl.kv[k] = finder.Get(k)
	}
	return nil
}

func TestProvisionerConfigure(t *testing.T) {
	log := florist.NewLogger("test")
	files := fstest.MapFS{"debug/dummyFile": {Data: []byte("A")}}
	secrets := fstest.MapFS{
		"base/unique": {Data: []byte("unique from base")},
		"base/f1":     {Data: []byte("f1 from base")},
		//
		"flowers/spy/f1": {Data: []byte("f1 from flowers")},
		"flowers/spy/f2": {Data: []byte("f2 from flowers")},
		"flowers/spy/f3": {Data: []byte("f3 from flowers")},
		//
		"nodes/x/spy/f2": {Data: []byte("f2 from nodes")},
		"nodes/x/spy/f4": {Data: []byte("f4 from nodes")},
	}

	want := map[string]string{
		"unique": "unique from base",
		"f1":     "f1 from flowers",
		"f2":     "f2 from nodes",
		"f3":     "f3 from flowers",
		"f4":     "f4 from nodes",
	}

	prov, err := provisioner.New(log, florist.CacheValidity, files, secrets)
	assert.NilError(t, err)

	prov.UseWorkdir()
	spy := &spyFlower{}
	err = prov.AddBouquet("x", "Stuff for node x", spy)
	assert.NilError(t, err)

	err = prov.Run([]string{"configure", "x"})
	assert.NilError(t, err)

	assert.DeepEqual(t, want, spy.kv)
}
