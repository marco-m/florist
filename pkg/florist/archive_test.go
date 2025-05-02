package florist_test

import (
	"io"
	"path/filepath"
	"testing"
	"time"

	"github.com/marco-m/florist/pkg/florist"
	"github.com/marco-m/rosina/assert"
)

func TestUnzipOne(t *testing.T) {
	florist.LowLevelInit(io.Discard, "Info", 24*time.Hour)
	dstDir := t.TempDir()

	type testCase struct {
		wantName string
	}

	test := func(t *testing.T, tc testCase) {
		dstPath := filepath.Join(dstDir, tc.wantName)

		err := florist.UnzipOne("testdata/archive/archive.zip",
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
	florist.LowLevelInit(io.Discard, "Info", 24*time.Hour)
	dstDir := t.TempDir()

	type testCase struct {
		wantName string
	}

	test := func(t *testing.T, tc testCase) {
		dstPath := filepath.Join(dstDir, tc.wantName)

		err := florist.UntarOne("testdata/archive/archive.tgz",
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

func TestUntarSome(t *testing.T) {
	florist.LowLevelInit(io.Discard, "Info", 24*time.Hour)
	dstDir := t.TempDir()
	owner, group, err := florist.WhoAmI()
	if err != nil {
		t.Fatalf("WhoAmI: %s", err)
	}

	tarPath := "testdata/archive/archive.tgz"
	some := []string{"file3", "file4"}
	err = florist.UntarSome(tarPath, dstDir, some, 0o644, owner, group)
	if err != nil {
		t.Fatalf("UntarAll: tar=%s, dst=%s: %s", tarPath, dstDir, err)
	}

	for _, fi := range []string{"file3", "file4"} {
		have := filepath.Join(dstDir, fi)
		want := filepath.Join("testdata/archive", fi)
		assert.FileEqualsFile(t, have, want)
	}
}

func TestUntarAll(t *testing.T) {
	florist.LowLevelInit(io.Discard, "Info", 24*time.Hour)
	dstDir := t.TempDir()
	owner, group, err := florist.WhoAmI()
	if err != nil {
		t.Fatalf("WhoAmI: %s", err)
	}

	tarPath := "testdata/archive/archive.tgz"
	err = florist.UntarAll(tarPath, dstDir, 0o644, owner, group)
	if err != nil {
		t.Fatalf("UntarAll: tar=%s, dst=%s: %s", tarPath, dstDir, err)
	}

	for _, fi := range []string{"file1.txt", "file2.txt", "file3", "file4"} {
		have := filepath.Join(dstDir, fi)
		want := filepath.Join("testdata/archive", fi)
		assert.FileEqualsFile(t, have, want)
	}
}
