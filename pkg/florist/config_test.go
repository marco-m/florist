package florist_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/marco-m/florist/pkg/florist"
	"github.com/marco-m/rosina/assert"
)

func TestConfigGetSuccess(t *testing.T) {
	cfg := setupConfig(t, t.TempDir(), map[string]string{"existing-key": "ciccio"})

	val := cfg.Get("existing-key")
	if have, want := val, "ciccio"; have != want {
		t.Errorf("have: %q; want: %q", have, want)
	}

	if err := cfg.Errors(); err != nil {
		t.Errorf("error: %s", err)
	}
}

func TestConfigGetFailure(t *testing.T) {
	cfg := setupConfig(t, t.TempDir(), map[string]string{"existing-key": "ciccio"})

	val1 := cfg.Get("non-existing-1")
	if have, want := val1, ""; have != want {
		t.Errorf("have: %q; want: %q", have, want)
	}

	val2 := cfg.Get("non-existing-2")
	if have, want := val2, ""; have != want {
		t.Errorf("have: %q; want: %q", have, want)
	}

	have := cfg.Errors().Error()
	want := "key 'non-existing-1': not found; key 'non-existing-2': not found"
	if !strings.Contains(have, want) {
		t.Errorf("\nhave: %q\ndoes not contain: %q", have, want)
	}
}

func TestConfigLookupSuccess(t *testing.T) {
	cfg := setupConfig(t, t.TempDir(), map[string]string{"existing-key": "ciccio"})

	val, err := cfg.Lookup("existing-key")
	if err != nil {
		t.Errorf("error: %s", err)
	}
	if have, want := val, "ciccio"; have != want {
		t.Errorf("have: %q; want: %q", have, want)
	}
}

func TestConfigLookupFailure(t *testing.T) {
	cfg := setupConfig(t, t.TempDir(), map[string]string{"existing-key": "ciccio"})

	_, err := cfg.Lookup("non-existing-1")
	want := "key 'non-existing-1': not found"
	if err == nil {
		t.Fatalf("error: <nil>; want: %s", want)
	}
	have := err.Error()
	if !strings.Contains(have, want) {
		t.Errorf("\nhave: %q\ndoes not contain: %q", have, want)
	}
}

func setupConfig(t *testing.T, dir string, kvs map[string]string) *florist.Config {
	t.Helper()

	file := filepath.Join(dir, "config.json")
	data, err := json.Marshal(kvs)
	assert.NoError(t, err, "Marshal")

	err = os.WriteFile(file, data, 0o664)
	assert.NoError(t, err, "WriteFile")

	cfg, err := florist.NewConfig(file)
	assert.NoError(t, err, "NewConfig")

	return cfg
}
