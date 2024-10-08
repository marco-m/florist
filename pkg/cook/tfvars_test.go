package cook

import (
	"strings"
	"testing"

	"github.com/marco-m/rosina"
	"gotest.tools/v3/assert"
	gotestfs "gotest.tools/v3/fs"
)

func TestParseTfVarsSuccess(t *testing.T) {
	type testCase struct {
		name  string
		input string
		want  map[string]string
	}

	run := func(t *testing.T, tc testCase) {
		have, err := parseTfVars(strings.NewReader(tc.input))

		rosina.AssertIsNil(t, err)
		rosina.AssertDeepEqual(t, have, tc.want, "tfvars")
	}

	testCases := []testCase{
		{
			name:  "one string",
			input: `a = "b"`,
			want:  map[string]string{"a": "b"},
		},
		{
			name:  "one int",
			input: `a = 42`,
			want:  map[string]string{"a": "42"},
		},
		{
			name:  "commented out",
			input: `# a = 42`,
			want:  map[string]string{},
		},
		{
			name: "two lines",
			input: `
a = 42
b = "a"`,
			want: map[string]string{"a": "42", "b": "a"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}

func TestParseTfVarsFailure(t *testing.T) {
	type testCase struct {
		name  string
		input string
		want  string
	}

	run := func(t *testing.T, tc testCase) {
		_, err := parseTfVars(strings.NewReader(tc.input))

		rosina.AssertIsNotNil(t, err)
		rosina.AssertContains(t, err.Error(), tc.want)
	}

	testCases := []testCase{
		{
			name: "duplicate",
			input: `
a = 42
b = "c"
a = 42
`,
			want: "parseTfVars: duplicate key: a",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}

func TestTfVarsToDirSuccess(t *testing.T) {
	dstDir := t.TempDir()

	err := TfVarsToDir("testdata/settings.tfvars", dstDir)
	rosina.AssertIsNil(t, err)

	mode := gotestfs.WithMode(0o640)
	want := gotestfs.Expected(t,
		gotestfs.WithMode(0o755),
		gotestfs.WithFile("location", "nbg1", mode),
		gotestfs.WithFile("ssh-port", "22", mode),
		gotestfs.WithFile("this_0", "foo", mode),
	)
	assert.Assert(t, gotestfs.Equal(dstDir, want))
}

func TestTfVarsToDirFailure(t *testing.T) {
	dstDir := t.TempDir()

	err := TfVarsToDir("testdata/duplicates.tfvars", dstDir)
	rosina.AssertIsNotNil(t, err)
	rosina.AssertContains(t, err.Error(),
		"TfVarsToDir: parseTfVars: duplicate key: a")
}
