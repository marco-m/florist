package florist_test

import (
	"path/filepath"
	"testing"

	"github.com/marco-m/florist/pkg/testhelpers"
	"github.com/marco-m/florist/pkg/florist"
	"github.com/marco-m/rosina/assert"
)

func TestUnzipOne(t *testing.T) {
	tempDir := t.TempDir()
	testhelpers.InitFlorist(t, tempDir)

	type testCase struct {
		wantName string
	}

	test := func(t *testing.T, tc testCase) {
		dstPath := filepath.Join(tempDir, tc.wantName)

		err := florist.UnzipOne("testdata/archive/two-files.zip",
			tc.wantName, dstPath)

		assert.NoError(t, err, "florist.UnzipOne")
		wantPath := filepath.Join("testdata/archive", tc.wantName)
		assert.FileEqualsFile(t, dstPath, wantPath)
	}

	testCases := []testCase{
		{wantName: "file1.txt"},
		{wantName: "file2.txt"},
	}

	for _, tc := range testCases {
		t.Run(tc.wantName, func(t *testing.T) { test(t, tc) })
	}
}

func TestUntarOne(t *testing.T) {
	tempDir := t.TempDir()
	testhelpers.InitFlorist(t, tempDir)

	type testCase struct {
		wantName string
	}

	test := func(t *testing.T, tc testCase) {
		dstPath := filepath.Join(tempDir, tc.wantName)

		err := florist.UntarOne("testdata/archive/two-files.tgz",
			tc.wantName, dstPath)

		assert.NoError(t, err, "florist.UntarOne")
		wantPath := filepath.Join("testdata/archive", tc.wantName)
		assert.FileEqualsFile(t, dstPath, wantPath)
	}

	testCases := []testCase{
		{wantName: "file1.txt"},
		{wantName: "file2.txt"},
	}

	for _, tc := range testCases {
		t.Run(tc.wantName, func(t *testing.T) { test(t, tc) })
	}
}

func TestUntarAll(t *testing.T) {
	tempDir := t.TempDir()
	testhelpers.InitFlorist(t, tempDir)

	owner, group, err := florist.WhoAmI()
	if err != nil {
		t.Fatalf("WhoAmI: %s", err)
	}

	tarPath := "testdata/archive/two-files.tgz"
	err = florist.UntarAll(tarPath, tempDir, 0o644, owner, group)
	if err != nil {
		t.Fatalf("UntarAll: tar=%s, dst=%s: %s", tarPath, tempDir, err)
	}

	for _, fi := range []string{"file1.txt", "file2.txt"} {
		have := filepath.Join(tempDir, fi)
		want := filepath.Join("testdata/archive", fi)
		assert.FileEqualsFile(t, have, want)
	}
}
