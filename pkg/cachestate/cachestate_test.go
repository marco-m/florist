package cachestate_test

import (
	"testing"
	"time"

	"github.com/marco-m/florist/internal"
	"github.com/marco-m/florist/pkg/cachestate"
)

func TestCacheStateIsNotValidBeforeUpdate(t *testing.T) {
	chache := cachestate.New(time.Hour, t.TempDir(), internal.MakeTestLog())

	have, want := chache.IsValid(), false
	if have != want {
		t.Errorf("\nCacheState.Isvalid: have: %v; want: %v", have, want)
	}
}

func TestCacheStateIsValidAfterUpdate(t *testing.T) {
	cache := cachestate.New(time.Hour, t.TempDir(), internal.MakeTestLog())

	err := cache.Update()
	if err != nil {
		t.Fatalf("\n%s (type: %T)", err, err)
	}

	have, want := cache.IsValid(), true
	if have != want {
		t.Errorf("\nCacheState.Isvalid: have: %v; want: %v", have, want)
	}
}

func TestCacheStateExpired(t *testing.T) {
	cache := cachestate.New(time.Millisecond, t.TempDir(), internal.MakeTestLog())

	err := cache.Update()
	if err != nil {
		t.Fatalf("\n%s (type: %T)", err, err)
	}

	// Validity too short: expired.
	time.Sleep(5 * time.Millisecond)
	have, want := cache.IsValid(), false
	if have != want {
		t.Errorf("\nCacheState.Isvalid too short: have: %v; want: %v", have, want)
	}
}

func TestFullLifeCycle(t *testing.T) {
	cache := cachestate.New(time.Hour, t.TempDir(), internal.MakeTestLog())

	err := cache.Update()
	if err != nil {
		t.Fatalf("\nUpdate 1: %s (type: %T)", err, err)
	}

	have, want := cache.IsValid(), true
	if have != want {
		t.Errorf("\nCacheState.Isvalid 1: have: %v; want: %v", have, want)
	}

	err = cache.Invalidate()
	if err != nil {
		t.Fatalf("\n%s (type: %T)", err, err)
	}

	have, want = cache.IsValid(), false
	if have != want {
		t.Errorf("\nCacheState.Isvalid 2: have: %v; want: %v", have, want)
	}

	err = cache.Update()
	if err != nil {
		t.Fatalf("\nUpdate 2: %s (type: %T)", err, err)
	}

	have, want = cache.IsValid(), true
	if have != want {
		t.Errorf("\nCacheState.Isvalid 3: have: %v; want: %v", have, want)
	}
}
