package installer_test

import (
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-hclog"
	"github.com/marco-m/florist"
	"github.com/marco-m/florist/flowers/test"
	"github.com/marco-m/florist/pkg/installer"

	"github.com/marco-m/xprog"
)

type MockOptions struct {
	Name string
}

type mockFlower struct {
	MockOptions
	Log hclog.Logger
}

func newMockFlower(opts MockOptions) (*mockFlower, error) {
	return &mockFlower{
		MockOptions: opts,
		Log:         hclog.NewNullLogger(),
	}, nil
}

func (fl *mockFlower) String() string {
	return fl.Name
}

func (fl *mockFlower) Description() string {
	return "I am a mock flower"
}

func (fl *mockFlower) Install() error {
	return nil
}

func TestInstallerAddOneFlowerSuccess(t *testing.T) {
	log := hclog.NewNullLogger()
	cacheValidity := 365 * 24 * time.Hour
	inst := installer.New(log, cacheValidity)
	flowers := []florist.Flower{&mockFlower{MockOptions: MockOptions{Name: "foo"}}}

	if err := inst.AddBouquet("", "", flowers...); err != nil {
		t.Fatal(err)
	}

	want := []installer.Bouquet{
		{
			Name:        "foo",
			Description: "I am a mock flower",
			Flowers:     flowers,
		},
	}

	if diff := cmp.Diff(want, inst.Bouquets()); diff != "" {
		t.Errorf("\nbouquets: mismatch (-want +have):\n%s", diff)
	}
}

func TestInstallerAddOneFlowerFailure(t *testing.T) {
	log := hclog.NewNullLogger()
	cacheValidity := 365 * 24 * time.Hour
	inst := installer.New(log, cacheValidity)
	bouquet := []florist.Flower{}
	wantErr := "AddBouquet: bouquet is empty"

	err := inst.AddBouquet("", "", bouquet...)

	if err == nil {
		t.Fatalf("have: <no error>; want: %s", wantErr)
	}
	if have := err.Error(); have != wantErr {
		t.Fatalf("have: %s; want: %s", have, wantErr)
	}
}

func TestInstallerAddMultipleFlowersSuccess(t *testing.T) {
	log := hclog.NewNullLogger()
	cacheValidity := 365 * 24 * time.Hour
	inst := installer.New(log, cacheValidity)

	flowers := []florist.Flower{
		&mockFlower{MockOptions: MockOptions{Name: "a"}},
		&mockFlower{MockOptions: MockOptions{Name: "b"}},
		&mockFlower{MockOptions: MockOptions{Name: "c"}},
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
	cacheValidity := 365 * 24 * time.Hour

	flowers := []florist.Flower{
		&mockFlower{MockOptions: MockOptions{Name: "a"}},
		&mockFlower{MockOptions: MockOptions{Name: "b"}},
		&mockFlower{MockOptions: MockOptions{Name: "c"}},
	}

	testCases := []struct {
		name         string
		bouquet      []florist.Flower
		bname        string
		bdescription string
		wantErr      string
	}{
		{
			name:    "more that one flower and name is empty",
			bouquet: flowers,
			wantErr: "AddBouquet: more that one flower and name is empty: [a b c]",
		},
		{
			name:    "more that one flower and desc is empty",
			bouquet: flowers,
			bname:   "foo",
			wantErr: "AddBouquet: more that one flower and description is empty: [a b c]",
		},
		{
			name: "xxx",
			bouquet: []florist.Flower{
				&mockFlower{MockOptions: MockOptions{Name: "a"}},
				&mockFlower{MockOptions: MockOptions{}},
			},
			bname:        "foo",
			bdescription: "bar",
			wantErr:      "AddBouquet: flower 1 has empty name",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			inst := installer.New(log, cacheValidity)

			err := inst.AddBouquet(tc.bname, tc.bdescription, tc.bouquet...)

			if err == nil {
				t.Fatalf("have: no error; want: %s", tc.wantErr)
			}
			have := err.Error()
			if have != tc.wantErr {
				t.Fatalf("\nhave: %s\nwant: %s", have, tc.wantErr)
			}
		})
	}
}

func TestInstallerDuplicateBouquetName(t *testing.T) {
	log := hclog.NewNullLogger()
	cacheValidity := 365 * 24 * time.Hour
	inst := installer.New(log, cacheValidity)

	bname := "pippo"
	bouquet1 := []florist.Flower{&mockFlower{MockOptions: MockOptions{Name: "1"}}}
	bouquet2 := []florist.Flower{&mockFlower{MockOptions: MockOptions{Name: "2"}}}
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

func TestInstallerRunVM(t *testing.T) {
	if xprog.Absent() {
		t.Skip("skip: test requires xprog")
	}

	flower := &test.Flower{
		Contents: "I am a little flower",
		Dst:      "/flowers/banana",
	}

	t.Run("installer runs successfully", func(t *testing.T) {
		log := hclog.NewNullLogger()
		inst := installer.New(log, florist.CacheValidityDefault)
		if err := inst.AddBouquet("", "", flower); err != nil {
			t.Fatal(err)
		}

		os.Args = []string{"sut", "install", "test"}
		err := inst.Run()

		if err != nil {
			t.Fatalf("\nhave: %s\nwant: <no error>", err)
		}
	})

	t.Run("can read what the flower wrote", func(t *testing.T) {
		buf, err := os.ReadFile(flower.Dst)
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(flower.Contents, string(buf)); diff != "" {
			t.Errorf("contents: mismatch (-want +have):\n%s", diff)
		}
	})
}
