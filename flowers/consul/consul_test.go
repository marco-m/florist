package consul_test

import (
	"embed"
	"io/fs"
	"testing"

	"github.com/gertd/wild"
	"github.com/google/go-cmp/cmp"
	"github.com/marco-m/florist"
	"github.com/marco-m/florist/flowers/consul"
	"github.com/marco-m/xprog"
)

//go:embed files
var filesFS embed.FS

func TestConsulServerInstallSuccessVM(t *testing.T) {
	if xprog.Absent() {
		t.Skip("skip: test requires xprog")
	}

	files, err := fs.Sub(filesFS, "files")
	if err != nil {
		t.Fatal("fs.Sub:", err)
	}
	florist.SetLogger(florist.NewLogger("test-consulserver"))
	cs, err := consul.NewServer(consul.ServerOptions{
		FilesFS: files,
		Version: "1.11.2",
		Hash:    "380eaff1b18a2b62d8e1d8a7cbc3f3e08b34d3f7187ee335b891ca2ba98784b3",
	})
	if err != nil {
		t.Fatal(err)
	}

	err = cs.Install()

	if err != nil {
		t.Fatalf("install:\nhave: %s\nwant: <no error>", err)
	}

}

func TestConsulServerInstallFailureVM(t *testing.T) {
	if xprog.Absent() {
		t.Skip("skip: test requires xprog")
	}

	// FIXME Since not compatible with a client install, I should wipe client installs before...
	// at this point, should I do it as a Flower method, Unistall(), or should I do it grossly only here?

	florist.SetLogger(florist.NewLogger("test-consulserver"))

	testCases := []struct {
		name        string
		opts        consul.ServerOptions
		wantErrWild string // wildcard matching
	}{
		{
			name:        "missing version",
			opts:        consul.ServerOptions{},
			wantErrWild: "* missing version",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := consul.NewServer(tc.opts)

			if err == nil {
				t.Fatalf("install:\nhave: <no error>\nwant: %s", tc.wantErrWild)
			}

			have := err.Error()
			if !wild.Match(tc.wantErrWild, have, false) {
				diff := cmp.Diff(tc.wantErrWild, have)
				t.Fatalf("error msg wildcard mismatch: (-want +have):\n%s", diff)
			}
		})
	}
}

func TestConsulClientInstallVM(t *testing.T) {
	if xprog.Absent() {
		t.Skip("skip: test requires xprog")
	}
	// FIXME WRITEME
}
