//go:build linux

package florist

import (
	"testing"

	"github.com/marco-m/rosina/assert"
)

func TestPrivateIpSuccess(t *testing.T) {
	have, err := PrivateIP("127.0.0.0/30")
	assert.NoError(t, err, "florist.InstanceIP")
	assert.Equal(t, have, "127.0.0.1", "florist.InstanceIP")
}

func TestPrivateIpFailure(t *testing.T) {
	type testCase struct {
		name        string
		networkCIDR string
		wantErr     string
	}

	test := func(t *testing.T, tc testCase) {
		_, err := PrivateIP(tc.networkCIDR)
		assert.ErrorContains(t, err, tc.wantErr)
	}

	testCases := []testCase{
		{
			name:        "invalid network CIDR",
			networkCIDR: "123.15.12.12/128",
			wantErr:     "invalid CIDR address",
		},
		{
			name:        "no match found",
			networkCIDR: "192.0.1.0/30",
			wantErr:     "none of the private IPs belongs to 192.0.1.0/30 network",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) { test(t, tc) })
	}
}
