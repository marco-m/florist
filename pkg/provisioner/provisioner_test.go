package provisioner_test

import (
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/marco-m/florist/pkg/florist"
	"github.com/marco-m/florist/pkg/provisioner"
)

func TestProvisionerConfigureZeroFlowers(t *testing.T) {
	var spy []string
	opts := &provisioner.Options{
		LogOutput: io.Discard,
		RootDir:   t.TempDir(),
		SetupFn: func(prov *provisioner.Provisioner) error {
			spy = append(spy, "SetupFn")
			return nil
		},
		PreConfigureFn: func(prov *provisioner.Provisioner, config *provisioner.Config) (any, error) {
			spy = append(spy, "PreConfigureFn")
			return nil, nil
		},
		PostConfigureFn: func(prov *provisioner.Provisioner, config *provisioner.Config, bag any) error {
			spy = append(spy, "PostConfigureFn")
			return nil
		},
	}
	cmdline := []string{"program", "configure", "--settings=testdata/simple.json"}
	err := provisioner.MainErr(cmdline, opts)
	if err != nil {
		t.Errorf("error: %s", err)
	}
	want := []string{"SetupFn", "PreConfigureFn", "PostConfigureFn"}
	if diff := cmp.Diff(want, spy); diff != "" {
		t.Errorf("spy mismatch:\n--- want\n+++ have\n%s", diff)
	}
}

type SpyFlower struct {
	Spy            *[]string
	Name           string
	InitError      error
	InstallError   error
	ConfigureError error
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
	*cc.Spy = append(*cc.Spy, fmt.Sprintf("SpyFlower.Init.%s.%s",
		cc.Name, stringErr(cc.InitError)))
	return cc.InitError
}

func (cc *SpyFlower) Install() error {
	*cc.Spy = append(*cc.Spy, fmt.Sprintf("SpyFlower.Install.%s.%s",
		cc.Name, stringErr(cc.InstallError)))
	return cc.InstallError
}

func (cc *SpyFlower) Configure() error {
	*cc.Spy = append(*cc.Spy, fmt.Sprintf("SpyFlower.Configure.%s.%s",
		cc.Name, stringErr(cc.ConfigureError)))
	return cc.ConfigureError
}

// Return a meaningful string also if err == nil.
func stringErr(err error) string {
	if err != nil {
		return err.Error()
	}
	return "<nil>"
}

func TestProvisionerConfigureTwoFlowers(t *testing.T) {
	var spy []string
	opts := &provisioner.Options{
		LogOutput: io.Discard,
		RootDir:   t.TempDir(),
		SetupFn: func(prov *provisioner.Provisioner) error {
			spy = append(spy, "SetupFn")
			return prov.AddFlowers(
				&SpyFlower{Spy: &spy, Name: "A"},
				&SpyFlower{Spy: &spy, Name: "B"},
			)
		},
		PreConfigureFn: func(prov *provisioner.Provisioner, config *provisioner.Config) (any, error) {
			spy = append(spy, "PreConfigureFn")
			return nil, nil
		},
		PostConfigureFn: func(prov *provisioner.Provisioner, config *provisioner.Config, bag any) error {
			spy = append(spy, "PostConfigureFn")
			return nil
		},
	}
	cmdline := []string{"program", "configure", "--settings=testdata/simple.json"}
	err := provisioner.MainErr(cmdline, opts)
	if err != nil {
		t.Errorf("error: %s", err)
	}
	want := []string{
		"SetupFn",
		"PreConfigureFn",
		"SpyFlower.Init.A.<nil>",
		"SpyFlower.Configure.A.<nil>",
		"SpyFlower.Init.B.<nil>",
		"SpyFlower.Configure.B.<nil>",
		"PostConfigureFn",
	}
	if diff := cmp.Diff(want, spy); diff != "" {
		t.Errorf("spy mismatch:\n--- want\n+++ have\n%s", diff)
	}
}

// This behavior changed in florist v0.5.0
// Before, it was terminating on first error.
func TestProvisionerIntermediateErrorsKeepsGoing(t *testing.T) {
	var spy []string
	opts := &provisioner.Options{
		LogOutput: io.Discard,
		RootDir:   t.TempDir(),
		SetupFn: func(prov *provisioner.Provisioner) error {
			spy = append(spy, "SetupFn")
			return prov.AddFlowers(
				&SpyFlower{
					Spy: &spy, Name: "A",
					InitError:      errors.New("E1"),
					ConfigureError: errors.New("E2"),
				},
				&SpyFlower{Spy: &spy, Name: "B"},
			)
		},
		PreConfigureFn: func(prov *provisioner.Provisioner, config *provisioner.Config) (any, error) {
			err := errors.New("E3")
			spy = append(spy, "PreConfigureFn."+err.Error())
			return nil, err
		},
		PostConfigureFn: func(prov *provisioner.Provisioner, config *provisioner.Config, bag any) error {
			spy = append(spy,
				"PostConfigureFn.<nil>",
				"PostConfigureFn.Provisioner.Errors."+stringErr(florist.JoinErrors(prov.Errors()...)),
			)
			return nil
		},
	}
	cmdline := []string{"program", "configure", "--settings=testdata/simple.json"}
	err := provisioner.MainErr(cmdline, opts)
	{
		have := err.Error()
		want := "configure: preconfigure: E3; flower init: E1; flower configure: E2"
		if have != want {
			t.Errorf("error mismatch:\nhave: %s\nwant: %s", have, want)
		}
	}
	{
		want := []string{
			"SetupFn",
			"PreConfigureFn.E3",
			"SpyFlower.Init.A.E1",
			"SpyFlower.Configure.A.E2",
			"SpyFlower.Init.B.<nil>",
			"SpyFlower.Configure.B.<nil>",
			"PostConfigureFn.<nil>",
			"PostConfigureFn.Provisioner.Errors.preconfigure: E3; flower init: E1; flower configure: E2",
		}
		if diff := cmp.Diff(want, spy); diff != "" {
			t.Errorf("spy mismatch:\n--- want\n+++ have\n%s", diff)
		}
	}
}
