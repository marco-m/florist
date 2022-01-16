package installer_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-hclog"
	"github.com/marco-m/florist"
	"github.com/marco-m/florist/pkg/installer"
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

func (fl *mockFlower) Install() error {
	return nil
}

func TestInstallerAddOneFlowerSuccess(t *testing.T) {
	log := hclog.NewNullLogger()
	cacheValidity := 365 * 24 * time.Hour
	inst := installer.New(log, cacheValidity)
	bouquet := []florist.Flower{&mockFlower{"foo"}}

	if err := inst.AddBouquet("", "", bouquet); err != nil {
		t.Fatal(err)
	}

	want := []installer.Bouquet{
		{
			Name:        "foo",
			Description: "I am a mock flower",
			Flowers:     []florist.Flower{&mockFlower{Name: "foo"}},
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

	err := inst.AddBouquet("", "", bouquet)

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

	bouquet := []florist.Flower{
		&mockFlower{"a"},
		&mockFlower{"b"},
		&mockFlower{"c"},
	}

	if err := inst.AddBouquet("pippo", "topolino", bouquet); err != nil {
		t.Fatal(err)
	}

	want := []installer.Bouquet{
		{
			Name:        "pippo",
			Description: "topolino",
			Flowers: []florist.Flower{
				&mockFlower{Name: "a"},
				&mockFlower{Name: "b"},
				&mockFlower{Name: "c"},
			},
		},
	}

	if diff := cmp.Diff(want, inst.Bouquets()); diff != "" {
		t.Errorf("\nlist: mismatch (-want +have):\n%s", diff)
	}
}

func TestInstallerAddMultipleFlowersFailure(t *testing.T) {
	log := hclog.NewNullLogger()
	cacheValidity := 365 * 24 * time.Hour

	testCases := []struct {
		name         string
		bouquet      []florist.Flower
		bname        string
		bdescription string
		wantErr      string
	}{
		{
			name: "more that one flower and name is empty",
			bouquet: []florist.Flower{
				&mockFlower{"a"}, &mockFlower{"b"}, &mockFlower{"c"},
			},
			wantErr: "AddBouquet: more that one flower and name is empty: [a b c]",
		},
		{
			name: "more that one flower and desc is empty",
			bouquet: []florist.Flower{
				&mockFlower{"a"}, &mockFlower{"b"}, &mockFlower{"c"},
			},
			bname:   "foo",
			wantErr: "AddBouquet: more that one flower and description is empty: [a b c]",
		},
		{
			name: "xxx",
			bouquet: []florist.Flower{
				&mockFlower{"a"}, &mockFlower{""}, &mockFlower{"c"},
			},
			bname:        "foo",
			bdescription: "bar",
			wantErr:      "AddBouquet: flower 1 has empty name",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			inst := installer.New(log, cacheValidity)

			err := inst.AddBouquet(tc.bname, tc.bdescription, tc.bouquet)

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
	bouquet1 := []florist.Flower{&mockFlower{"1"}}
	bouquet2 := []florist.Flower{&mockFlower{"2"}}
	wantErr := "AddBouquet: there is already a bouquet with name pippo"

	if err := inst.AddBouquet(bname, "topolino", bouquet1); err != nil {
		t.Fatalf("have: %s; want: <no error>", err)
	}

	err := inst.AddBouquet(bname, "clarabella", bouquet2)

	have := "<no error>"
	if err != nil {
		have = err.Error()
	}
	if have != wantErr {
		t.Fatalf("\nhave: %s\nwant: %s", have, wantErr)
	}
}
	}
}
