// Running gopass tests requires too much user-specific setup (for example, PGP key).
// I could protect these tests with florist.SkipIfNotDisposableHost(t) but also in that
// case I would have to set up everything and I don't have time now.
//
// Instead, we run these tests only if we find gopass already installed on the host.
// Not perfect but good enough.
//
// gopass init --storage fs --crypto age

package cook_test

import (
	"math/rand"
	"os/exec"
	"strconv"
	"testing"
	"time"

	"github.com/marco-m/florist/pkg/cook"
	"gotest.tools/v3/assert"
)

func TestGopassDeleteNonExistingKey(t *testing.T) {
	skipIfGopassNotFound(t)

	rand.Seed(time.Now().Unix())
	key := "cook/" + strconv.FormatUint(rand.Uint64(), 16)

	err := cook.GopassDelete(key)

	assert.NilError(t, err)
}

func TestGopassDeleteExistingKey(t *testing.T) {
	skipIfGopassNotFound(t)

	// insert
	key, val := "cook/k3", "v3"
	assert.NilError(t, cook.GopassPut(key, val))

	// delete
	err := cook.GopassDelete(key)

	assert.NilError(t, err)
	_, err = cook.GopassGet(key)
	assert.ErrorIs(t, err, cook.ErrNotFound)
}

func TestGopassGetExistingKey(t *testing.T) {
	skipIfGopassNotFound(t)

	key, val := "cook/k1", "v1"
	err := cook.GopassPut(key, val)
	assert.NilError(t, err)

	have, err := cook.GopassGet(key)

	assert.NilError(t, err)
	assert.Equal(t, have, val)

	// Cleanup
	assert.NilError(t, cook.GopassDelete(key))
}

func TestGopassGetNonExistingKey(t *testing.T) {
	skipIfGopassNotFound(t)

	key := "cook/k1"
	assert.NilError(t, cook.GopassDelete(key))

	have, err := cook.GopassGet(key)

	assert.ErrorIs(t, err, cook.ErrNotFound)
	assert.Equal(t, have, "")
}

func TestGopassPutSuccess(t *testing.T) {
	skipIfGopassNotFound(t)

	key, val := "cook/k2", "v2"
	err := cook.GopassDelete(key)
	assert.NilError(t, err)

	err = cook.GopassPut(key, val)
	assert.NilError(t, err)

	have, err := cook.GopassGet(key)
	assert.NilError(t, err)
	assert.Equal(t, have, val)
}

func skipIfGopassNotFound(t *testing.T) {
	t.Helper()
	_, err := exec.LookPath("gopass")
	if err != nil {
		t.Skip("skip: gopass not found")
	}
}
