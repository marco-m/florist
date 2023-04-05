package florist_test

import (
	"testing"
	"testing/fstest"

	"gotest.tools/v3/assert"

	"github.com/marco-m/florist/pkg/florist"
)

var fsys = fstest.MapFS{
	"f1":   {Data: []byte("1")},
	"d/f2": {Data: []byte("2")},
}

func TestFsFinderOneGetSuccess(t *testing.T) {
	finder := florist.NewFsFinder(fsys)

	// If you need a single value, it is better to use Lookup instead of Get.
	// Here we use Get only to test.

	have := finder.Get("f1")
	assert.Equal(t, have, "1")

	err := finder.Error()
	assert.NilError(t, err)
}

func TestFsFinderMultipleGetSuccess(t *testing.T) {
	finder := florist.NewFsFinder(fsys)

	// This is the typical usage of Get, in a struct.
	data := struct {
		v1 string
		v2 string
	}{
		v1: finder.Get("f1"),
		v2: finder.Get("d/f2"),
	}

	assert.Equal(t, data.v1, "1")
	assert.Equal(t, data.v2, "2")

	err := finder.Error()
	assert.NilError(t, err)
}

func TestFsFinderOneLookupSuccess(t *testing.T) {
	finder := florist.NewFsFinder(fsys)

	// If you need a single value, it is better to use Lookup instead of Get.

	have, err := finder.Lookup("f1")

	assert.Equal(t, have, "1")
	assert.NilError(t, err)
}

func TestFsFinderGetMultipleFailure(t *testing.T) {
	finder := florist.NewFsFinder(fsys)

	data := struct {
		v1 string
		v2 string
	}{
		v1: finder.Get("x"),
		v2: finder.Get("y"),
	}

	assert.Equal(t, data.v1, "")
	assert.Equal(t, data.v2, "")

	err := finder.Error()
	assert.ErrorContains(t, err, "multiple errors")
}

var fsysA = fstest.MapFS{
	"f1": {Data: []byte("1")}, // Overrides entry in fsysB
}

var fsysB = fstest.MapFS{
	"f1":   {Data: []byte("2")},
	"d/f2": {Data: []byte("3")},
}

func TestUnionFinderSuccess(t *testing.T) {
	finderA := florist.NewFsFinder(fsysA)
	finderB := florist.NewFsFinder(fsysB)
	finder := florist.NewUnionFinder(finderA, finderB)

	data := struct {
		v1 string
		v2 string
	}{
		v1: finder.Get("f1"),   // Found in A (overrides B).
		v2: finder.Get("d/f2"), // Not found in A, found in B.
	}
	assert.Equal(t, data.v1, "1")
	assert.Equal(t, data.v2, "3")

	err := finder.Error()
	assert.NilError(t, err)
}

func TestUnionFinderFailure(t *testing.T) {
	finderA := florist.NewFsFinder(fsysA)
	finderB := florist.NewFsFinder(fsysB)
	finder := florist.NewUnionFinder(finderA, finderB)

	have, err := finder.Lookup("x")
	assert.Equal(t, have, "")
	assert.Error(t, err, "FsFinder.Get: open x: file does not exist")
}
