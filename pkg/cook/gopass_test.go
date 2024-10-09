// Running gopass tests requires too much user-specific setup (for example, PGP
// key). We could protect these tests with florist.SkipIfNotDisposableHost(t)
// but also in that case we would have to set up everything and I don't have
// time now.
//
// Instead, we run these tests only if we find gopass already installed on the
// host. Not perfect but good enough.
//
// gopass init --storage fs --crypto age

package cook_test

import (
	"fmt"
	"io/fs"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"

	"gotest.tools/v3/assert"
	gotestfs "gotest.tools/v3/fs"

	"github.com/marco-m/florist/pkg/cook"
	"github.com/marco-m/rosina"
)

func TestGopassDeleteNonExistingKey(t *testing.T) {
	skipIfGopassNotFound(t)

	key := "cook/" + strconv.FormatUint(rand.Uint64(), 16)

	err := cook.GopassDelete(key)

	rosina.AssertNoError(t, err)
}

func TestGopassDeleteExistingKey(t *testing.T) {
	skipIfGopassNotFound(t)

	// insert
	key, val := "cook/k3", "v3"
	rosina.AssertNoError(t, cook.GopassPut(key, val))

	// delete
	err := cook.GopassDelete(key)

	rosina.AssertNoError(t, err)
	_, err = cook.GopassGet(key)
	rosina.AssertErrorIs(t, err, cook.ErrNotFound)
}

func TestGopassGetExistingKey(t *testing.T) {
	skipIfGopassNotFound(t)

	key, val := "cook/k1", "v1"
	err := cook.GopassPut(key, val)
	rosina.AssertNoError(t, err)

	have, err := cook.GopassGet(key)

	rosina.AssertNoError(t, err)
	rosina.AssertEqual(t, have, val, "value")

	// Cleanup
	rosina.AssertNoError(t, cook.GopassDelete(key))
}

func TestGopassGetNonExistingKey(t *testing.T) {
	skipIfGopassNotFound(t)

	key := "cook/k1"
	rosina.AssertNoError(t, cook.GopassDelete(key))

	have, err := cook.GopassGet(key)

	rosina.AssertErrorIs(t, err, cook.ErrNotFound)
	rosina.AssertEqual(t, have, "", "value")
}

func TestGopassPutSuccess(t *testing.T) {
	skipIfGopassNotFound(t)

	key, val := "cook/k2", "v2"
	err := cook.GopassDelete(key)
	rosina.AssertNoError(t, err)

	err = cook.GopassPut(key, val)
	rosina.AssertNoError(t, err)

	have, err := cook.GopassGet(key)
	rosina.AssertNoError(t, err)
	rosina.AssertEqual(t, have, val, "value")
}

func TestGopassLsSuccess(t *testing.T) {
	skipIfGopassNotFound(t)

	want := []string{
		"a/b/k3",
		"a/k2",
		"k1",
	}

	have, err := cook.GopassLs("florist.test/secrets")
	rosina.AssertNoError(t, err)
	rosina.AssertDeepEqual(t, have, want, "values")
}

func TestGopassLsFailure(t *testing.T) {
	skipIfGopassNotFound(t)

	_, err := cook.GopassLs("florist.test/non-existing")
	rosina.AssertErrorContains(t, err, "not found")
}

// gopass ls florist.test
// florist.test/
// ├── outside/   <= everything else will not be exported
// │   └── q
// └── secrets/   <= we point to this directory
//     ├── a/
//     │   ├── b/
//     │   │   └── k3
//     │   └── k2
//     └── k1

func TestGopassToDirSuccess(t *testing.T) {
	skipIfGopassNotFound(t)

	dstDir := filepath.Join(t.TempDir(), "secrets")

	err := cook.GopassToDir("florist.test/secrets", dstDir)
	rosina.AssertNoError(t, err)

	want := gotestfs.Expected(t,
		gotestfs.WithMode(0o755),
		gotestfs.WithFile("k1", "v1", gotestfs.WithMode(0o600)),
		gotestfs.WithDir("a", gotestfs.WithMode(0o755),
			gotestfs.WithFile("k2", "v2", gotestfs.WithMode(0o600)),
			gotestfs.WithDir("b", gotestfs.WithMode(0o755),
				gotestfs.WithFile("k3", "v3", gotestfs.WithMode(0o600)),
			),
		),
	)
	// repr.Println("manifest", want)
	rosina.AssertNoError(t, walkDir(dstDir))
	assert.Assert(t, gotestfs.Equal(dstDir, want))
}

func walkDir(root string) error {
	fn := func(path string, d fs.DirEntry, err error) error {
		fmt.Println("=>", path)
		return nil
	}

	return filepath.WalkDir(root, fn)
}

func skipIfGopassNotFound(t *testing.T) {
	t.Helper()
	if os.Getenv("FLORIST_TEST_GOPASS") == "" {
		t.Skip("skip: FLORIST_TEST_GOPASS not set")
	}
	_, err := exec.LookPath("gopass")
	if err != nil {
		t.Skip("skip: gopass not found")
	}
}
