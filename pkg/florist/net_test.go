package florist_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/marco-m/florist/pkg/florist"
	"github.com/marco-m/florist/pkg/provisioner"
	"github.com/marco-m/rosina/assert"
)

func TestNetFetchMockSuccess(t *testing.T) {
	err := provisioner.LowLevelInit(io.Discard, "INFO")
	assert.NoError(t, err, "provisioner.LowLevelInit")
	dir := t.TempDir()
	hash := "b493d48364afe44d11c0165cf470a4164d1e2609911ef998be868d46ade3de4e"
	client := &http.Client{Timeout: 1 * time.Second}
	contents := "banana"

	ts := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, contents)
		}))
	defer ts.Close()

	path, err := florist.NetFetch(client, ts.URL, florist.SHA256, hash, dir)
	assert.NoError(t, err, "florist.NetFetch")
	assert.FileEqualsString(t, path, contents)
}

func TestNetFetchMockFailure(t *testing.T) {
	testCases := []struct {
		name    string
		hash    string
		handler func(w http.ResponseWriter, r *http.Request)
		wantErr string
	}{
		{
			name: "Not found",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			},
			wantErr: `NetFetch: received 404 Not Found (GET http://127.0.0.1:`,
		},
		{
			name: "hash mismatch",
			hash: "b493d48364afe44d11c0165cf470a4164d1e2609911ef998be868d46ade3de4e",
			handler: func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, "foobar")
			},
			wantErr: "NetFetch: hash mismatch: have: c3ab8ff13720e8ad9047dd39466b3c8974e592c2fa383d4a3960714caef0c4f2; want: b493d48364afe44d11c0165cf470a4164d1e2609911ef998be868d46ade3de4e",
		},
	}

	dir := t.TempDir()
	client := &http.Client{Timeout: 1 * time.Second}

	for tcN, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(tc.handler))
			defer ts.Close()

			url := fmt.Sprintf("%s/%s", ts.URL, strconv.Itoa(tcN))
			_, err := florist.NetFetch(client, url, florist.SHA256, tc.hash, dir)

			assert.ErrorContains(t, err, tc.wantErr)
		})
	}
}
