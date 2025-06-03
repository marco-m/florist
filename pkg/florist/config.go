package florist

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	settings     map[string]string
	errs         []string
	settingsPath string
}

func NewConfig(settingsPath string) (*Config, error) {
	cfg := &Config{settingsPath: settingsPath}
	if err := parse(cfg.settingsPath, &cfg.settings); err != nil {
		return nil, fmt.Errorf("NewConfig: parsing %s: %s", cfg.settingsPath, err)
	}
	return cfg, nil
}

func parse(path string, data any) error {
	rd, err := os.Open(path)
	if err != nil {
		return err
	}
	defer rd.Close()
	dec := json.NewDecoder(rd)
	if err := dec.Decode(data); err != nil {
		return err
	}
	return nil
}

// Get returns the value of key k if found. If the key is missing, it returns
// the empty string and adds the error to the list returned by Errors.
// This allows a simple sequence of calling Get multiple times and checking for
// all the keys that were missing keys only once at the end, by calling Errors.
//
// If on the other end you want to know immediately if the key is missing, use
// Lookup.
func (cfg *Config) Get(k string) string {
	v, found := cfg.settings[k]

	if !found {
		cfg.errs = append(cfg.errs, fmt.Sprintf("key '%s': not found", k))
	}
	return v
}

// Lookup returns the value of key k if found. If the key is missing, it returns
// an error. Contrary to Get, it does not append a lookup failure to the errors
// returned by Errors.
func (cfg *Config) Lookup(k string) (string, error) {
	v, found := cfg.settings[k]
	if !found {
		return "", fmt.Errorf("key '%s': not found (file: %s)", k, cfg.settingsPath)
	}
	return v, nil
}

func (cfg *Config) Errors() error {
	if len(cfg.errs) > 0 {
		return fmt.Errorf("%s (file: %s)", strings.Join(cfg.errs, "; "),
			cfg.settingsPath)
	}
	return nil
}
