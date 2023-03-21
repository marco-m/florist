package installer_test

import (
	"io/fs"
	"testing"

	"github.com/hashicorp/go-hclog"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"

	"github.com/marco-m/florist"
	"github.com/marco-m/florist/pkg/installer"
)

type mockFlower struct {
	Name string
	Log  hclog.Logger
}

func (fl *mockFlower) String() string {
	return fl.Name
}

func (fl *mockFlower) Description() string {
	return "I am a mock flower"
}

func (fl *mockFlower) Init() error {
	if fl.Log == nil {
		fl.Log = hclog.NewNullLogger()
	}
	return nil
}

func (fl *mockFlower) Install(files fs.FS, finder florist.Finder) error {
	return nil
}

func (fl *mockFlower) Configure(files fs.FS, finder florist.Finder) error {
	return nil
}

func TestInstallerAddBouquetSuccess(t *testing.T) {
	log := hclog.NewNullLogger()
	inst, err := installer.New(log, florist.CacheValidity, nil, nil)
	assert.NilError(t, err)

	flowers := []florist.Flower{
		&mockFlower{Name: "a"},
		&mockFlower{Name: "b"},
		&mockFlower{Name: "c"},
	}

	err = inst.AddBouquet("goofy", "mickey's friend", flowers...)
	assert.NilError(t, err)

	have := inst.Bouquets()
	want := []installer.Bouquet{
		{
			Name:        "goofy",
			Description: "mickey's friend",
			Flowers:     flowers,
		},
	}
	assert.Assert(t, cmp.DeepEqual(have, want))
}

func TestInstallerAddBouquetFailure(t *testing.T) {
	log := hclog.NewNullLogger()

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
			inst, err := installer.New(log, florist.CacheValidity, nil, nil)
			assert.NilError(t, err)

			err = inst.AddBouquet(tc.bname, tc.bdescription, tc.bouquet...)

			assert.ErrorContains(t, err, tc.wantErr)
		})
	}
}

func TestInstallerDuplicateBouquetName(t *testing.T) {
	log := hclog.NewNullLogger()
	inst, err := installer.New(log, florist.CacheValidity, nil, nil)
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
