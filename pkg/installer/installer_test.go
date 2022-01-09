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

func TestInstallerAddOneFlower(t *testing.T) {
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
	inst := installer.New(log, cacheValidity)

	bouquet := []florist.Flower{
		&mockFlower{"a"},
		&mockFlower{"b"},
		&mockFlower{"c"},
	}

	err := inst.AddBouquet("", "", bouquet)
	if err == nil {
		t.Fatal("have: no error; want: error")
	}
	have := err.Error()
	want := "AddBouquet: more that one flower and name is empty: [a b c]"
	if have != want {
		t.Fatalf("have: %s; want: %s", have, want)
	}
}
