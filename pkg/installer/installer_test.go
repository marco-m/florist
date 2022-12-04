package installer_test

import (
	"os"
	"path"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-hclog"
	"github.com/marco-m/florist"
	"github.com/marco-m/florist/flowers/test"
	"github.com/marco-m/florist/pkg/installer"
	"gotest.tools/v3/assert"
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

func (fl *mockFlower) Install() error {
	return nil
}

func (fl *mockFlower) Configure(rawCfg []byte) error {
	return nil
}

func TestInstallerAddFlowerSuccess(t *testing.T) {
	log := hclog.NewNullLogger()
	inst := installer.New(log, florist.CacheValidityDefault, nil)
	flower := &mockFlower{Name: "foo"}

	if err := inst.AddFlower(flower); err != nil {
		t.Fatal(err)
	}

	want := []installer.Bouquet{
		{
			Name:        "foo",
			Description: "I am a mock flower",
			Flowers:     []florist.Flower{flower},
		},
	}

	if diff := cmp.Diff(want, inst.Bouquets()); diff != "" {
		t.Errorf("\nbouquets: mismatch (-want +have):\n%s", diff)
	}
}

func TestInstallerAddFlowerFailure(t *testing.T) {
	log := hclog.NewNullLogger()
	inst := installer.New(log, florist.CacheValidityDefault, nil)
	flower := &mockFlower{Name: ""}

	err := inst.AddFlower(flower)

	assert.ErrorContains(t, err, "AddBouquet: name cannot be empty")
}

func TestInstallerAddBouquetSuccess(t *testing.T) {
	log := hclog.NewNullLogger()
	inst := installer.New(log, florist.CacheValidityDefault, nil)

	flowers := []florist.Flower{
		&mockFlower{Name: "a"},
		&mockFlower{Name: "b"},
		&mockFlower{Name: "c"},
	}

	if err := inst.AddBouquet("pippo", "topolino", flowers...); err != nil {
		t.Fatal(err)
	}

	want := []installer.Bouquet{
		{
			Name:        "pippo",
			Description: "topolino",
			Flowers:     flowers,
		},
	}

	if diff := cmp.Diff(want, inst.Bouquets()); diff != "" {
		t.Errorf("\nlist: mismatch (-want +have):\n%s", diff)
	}
}

func TestInstallerAddMultipleFlowersFailure(t *testing.T) {
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
			inst := installer.New(log, florist.CacheValidityDefault, nil)

			err := inst.AddBouquet(tc.bname, tc.bdescription, tc.bouquet...)

			assert.ErrorContains(t, err, tc.wantErr)
		})
	}
}

func TestInstallerDuplicateBouquetName(t *testing.T) {
	log := hclog.NewNullLogger()
	inst := installer.New(log, florist.CacheValidityDefault, nil)

	bname := "pippo"
	bouquet1 := []florist.Flower{&mockFlower{Name: "1"}}
	bouquet2 := []florist.Flower{&mockFlower{Name: "2"}}
	wantErr := "AddBouquet: there is already a bouquet with name pippo"

	if err := inst.AddBouquet(bname, "topolino", bouquet1...); err != nil {
		t.Fatalf("have: %s; want: <no error>", err)
	}

	err := inst.AddBouquet(bname, "clarabella", bouquet2...)

	have := "<no error>"
	if err != nil {
		have = err.Error()
	}
	if have != wantErr {
		t.Fatalf("\nhave: %s\nwant: %s", have, wantErr)
	}
}

func TestInstall(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)

	flower := &test.Flower{
		Contents: "I am a little flower",
		DstPath:  "/flowers/banana",
	}

	t.Run("install runs successfully", func(t *testing.T) {
		log := hclog.NewNullLogger()
		inst := installer.New(log, florist.CacheValidityDefault, nil)
		assert.NilError(t, inst.AddFlower(flower))

		os.Args = []string{"sut", "install", "test"}
		assert.NilError(t, inst.Run())
	})

	t.Run("can read what the install step wrote", func(t *testing.T) {
		buf, err := os.ReadFile(flower.DstPath)
		assert.NilError(t, err)
		have := string(buf)
		assert.Equal(t, have, flower.Contents)
	})
}

func TestConfigure(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)

	dir := t.TempDir()
	flower := &test.Flower{
		SrcPath: "testdata/flower.txt.tmpl",
		DstPath: path.Join(dir, "flower.txt"),
	}

	// TODO one secres file for two flowers...

	t.Run("configure runs successfully", func(t *testing.T) {
		log := hclog.NewNullLogger()
		inst := installer.New(log, florist.CacheValidityDefault, nil)
		assert.NilError(t, inst.AddFlower(flower))

		os.Args = []string{"sut", "configure",
			"--cfg=" + "testdata/florist.config.json", "test"}
		assert.NilError(t, inst.Run())
	})

	t.Run("can read what the configure step wrote", func(t *testing.T) {
		buf, err := os.ReadFile(flower.DstPath)
		assert.NilError(t, err)
		have := string(buf)
		want := "I am a little strawberry"
		assert.Equal(t, have, want)
	})
}
