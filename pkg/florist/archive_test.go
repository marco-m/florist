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
	florist.LowLevelInit(io.Discard, "Info", 24*time.Hour)
	dstDir := t.TempDir()

	type testCase struct {
		wantName string
	}

	test := func(t *testing.T, tc testCase) {
		dstPath := filepath.Join(dstDir, tc.wantName)

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
