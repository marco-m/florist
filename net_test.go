package florist_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/gertd/wild"
	"github.com/google/go-cmp/cmp"
	"github.com/marco-m/florist"
)

func TestNetFetchMockSuccess(t *testing.T) {
	dir, err := os.MkdirTemp("", "test-netfetch")
	if err != nil {
		t.Fatal("tmpdir:", err)
	}
	defer func() { os.RemoveAll(dir) }()

	hash := "b493d48364afe44d11c0165cf470a4164d1e2609911ef998be868d46ade3de4e"
	client := &http.Client{Timeout: 1 * time.Second}
	contents := "banana"

	ts := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, contents)
		}))
	defer ts.Close()

	path, err := florist.NetFetch(client, ts.URL, florist.SHA256, hash, dir)

	if err != nil {
		t.Fatalf("\nhave: %s\nwant: <no error>", err)
	}

	buf, err := os.ReadFile(path)
	if err != nil {
		t.Fatal("readfile:", err)
	}
	if diff := cmp.Diff(contents, string(buf)); diff != "" {
		t.Fatalf("file contents mismatch: (-want +have):\n%s", diff)
	}
}

func TestNetFetchMockFailure(t *testing.T) {
	testCases := []struct {
		name        string
		hash        string
		handler     func(w http.ResponseWriter, r *http.Request)
		wantErrWild string // wildcard matching
	}{
		{
			name: "Not found",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			},
			wantErrWild: `NetFetch: received 404 Not Found (GET http://127.0.0.1:*/0)`,
		},
		{
			name: "hash mismatch",
			hash: "b493d48364afe44d11c0165cf470a4164d1e2609911ef998be868d46ade3de4e",
			handler: func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, "foobar")
			},
			wantErrWild: "NetFetch: hash mismatch: have: c3ab8ff13720e8ad9047dd39466b3c8974e592c2fa383d4a3960714caef0c4f2; want: b493d48364afe44d11c0165cf470a4164d1e2609911ef998be868d46ade3de4e",
		},
	}

	dir, err := os.MkdirTemp("", "test-netfetch")
	if err != nil {
		t.Fatal("tmpdir:", err)
	}
	defer func() { os.RemoveAll(dir) }()

	client := &http.Client{Timeout: 1 * time.Second}

	for tcN, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(tc.handler))
			defer ts.Close()

			url := fmt.Sprintf("%s/%s", ts.URL, strconv.Itoa(tcN))
			_, err := florist.NetFetch(client, url, florist.SHA256, tc.hash, dir)

			if err == nil {
				t.Fatalf("\nhave: <no error>\nwant: %s", tc.wantErrWild)
			}

			have := err.Error()
			if !wild.Match(tc.wantErrWild, have, false) {
				diff := cmp.Diff(tc.wantErrWild, have)
				t.Fatalf("error msg wildcard mismatch: (-want +have):\n%s", diff)
			}
		})
	}
}
