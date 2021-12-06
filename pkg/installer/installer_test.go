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
	name string
}

func (fl *mockFlower) Install() error {
	return nil
}

func (fl *mockFlower) SetLogger(log hclog.Logger) {}

func (fl *mockFlower) Description() florist.Description {
	return florist.Description{Name: fl.name, Long: "I am a mock flower"}

}

func TestInstallerAddSimple(t *testing.T) {
	log := hclog.NewNullLogger()
	cacheValidity := 365 * 24 * time.Hour
	inst := installer.New(log, cacheValidity)

	fooFlower := mockFlower{"foo"}

	if err := inst.AddFlower(&fooFlower); err != nil {
		t.Fatal(err)
	}

	want := [][]string{{"foo"}}

	if diff := cmp.Diff(want, inst.List()); diff != "" {
		t.Errorf("\nlist: mismatch (-want +have):\n%s", diff)
	}
}

func TestInstallerAddSubFlowers(t *testing.T) {
	log := hclog.NewNullLogger()
	cacheValidity := 365 * 24 * time.Hour
	inst := installer.New(log, cacheValidity)

	af := mockFlower{"a"}
	bf := mockFlower{"b"}
	cf := mockFlower{"c"}

	if err := inst.AddFlower(&af); err != nil {
		t.Fatal(err)
	}
	if err := inst.AddFlower(&bf); err != nil {
		t.Fatal(err)
	}
	if err := inst.AddFlower(&cf, "a", "b"); err != nil {
		t.Fatal(err)
	}

	want := [][]string{{"a"}, {"b"}, {"c", "a", "b"}}

	if diff := cmp.Diff(want, inst.List()); diff != "" {
		t.Errorf("\nlist: mismatch (-want +have):\n%s", diff)
	}
}
