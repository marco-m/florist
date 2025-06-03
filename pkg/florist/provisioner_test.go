package florist_test

import (
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/marco-m/florist/pkg/florist"
)

func TestProvisionerConfigureZeroFlowers(t *testing.T) {
	var spy []string
	opts := &florist.Options{
		LogOutput: io.Discard,
		RootDir:   t.TempDir(),
		SetupFn: func(prov *florist.Provisioner) error {
			spy = append(spy, "SetupFn")
			return nil
		},
		PreConfigureFn: func(prov *florist.Provisioner, config *florist.Config) (any, error) {
			spy = append(spy, "PreConfigureFn")
			return nil, nil
		},
		PostConfigureFn: func(prov *florist.Provisioner, config *florist.Config, bag any) error {
			spy = append(spy, "PostConfigureFn")
			return nil
		},
	}
	cmdline := []string{"program", "configure", "--settings=testdata/simple.json"}
	err := florist.MainErr(cmdline, opts)
	if err != nil {
		t.Errorf("error: %s", err)
	}
	want := []string{"SetupFn", "PreConfigureFn", "PostConfigureFn"}
	if diff := cmp.Diff(want, spy); diff != "" {
		t.Errorf("spy mismatch:\n--- want\n+++ have\n%s", diff)
	}
}

type SpyFlower struct {
	Spy  *[]string
	Name string
}

func (cc *SpyFlower) String() string {
	return "SpyFlower:" + cc.Name
}

func (cc *SpyFlower) Description() string {
	return "The Spy Flower " + cc.Name
}

func (cc *SpyFlower) Embedded() []string {
	return nil
}

func (cc *SpyFlower) Init() error {
	*cc.Spy = append(*cc.Spy, "SpyFlower.Init."+cc.Name)
	return nil
}

func (cc *SpyFlower) Install() error {
	*cc.Spy = append(*cc.Spy, "SpyFlower.Install."+cc.Name)
	return nil
}

func (cc *SpyFlower) Configure() error {
	*cc.Spy = append(*cc.Spy, "SpyFlower.Configure."+cc.Name)
	return nil
}

func TestProvisionerConfigureTwoFlowers(t *testing.T) {
	var spy []string
	opts := &florist.Options{
		LogOutput: io.Discard,
		RootDir:   t.TempDir(),
		SetupFn: func(prov *florist.Provisioner) error {
			spy = append(spy, "SetupFn")
			return prov.AddFlowers(
				&SpyFlower{Spy: &spy, Name: "A"},
				&SpyFlower{Spy: &spy, Name: "B"},
			)
		},
		PreConfigureFn: func(prov *florist.Provisioner, config *florist.Config) (any, error) {
			spy = append(spy, "PreConfigureFn")
			return nil, nil
		},
		PostConfigureFn: func(prov *florist.Provisioner, config *florist.Config, bag any) error {
			spy = append(spy, "PostConfigureFn")
			return nil
		},
	}
	cmdline := []string{"program", "configure", "--settings=testdata/simple.json"}
	err := florist.MainErr(cmdline, opts)
	if err != nil {
		t.Errorf("error: %s", err)
	}
	want := []string{
		"SetupFn",
		"PreConfigureFn",
		"SpyFlower.Init.A",
		"SpyFlower.Configure.A",
		"SpyFlower.Init.B",
		"SpyFlower.Configure.B",
		"PostConfigureFn",
	}
	if diff := cmp.Diff(want, spy); diff != "" {
		t.Errorf("spy mismatch:\n--- want\n+++ have\n%s", diff)
	}
}
