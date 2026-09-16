package secrets

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Store persists resolved MCP configuration values to one untracked,
// workspace-local file, never to the generated team index, .miez/miez.lock.yaml,
// or .miez/miez_modules/.
type Store struct {
	Path string
}

// Load reads previously supplied values, returning an empty map when the
// file does not yet exist.
func (store Store) Load() (map[string]string, error) {
	data, err := os.ReadFile(store.Path)
	if errors.Is(err, os.ErrNotExist) {
		return map[string]string{}, nil
	}
	if err != nil {
		return nil, err
	}
	values := map[string]string{}
	if err := yaml.Unmarshal(data, &values); err != nil {
		return nil, err
	}
	return values, nil
}

// Merge loads the existing store, applies additions on top, and saves the
// result, so a second render reads previously supplied values back.
func (store Store) Merge(additions map[string]string) error {
	values, err := store.Load()
	if err != nil {
		return err
	}
	for key, value := range additions {
		values[key] = value
	}
	data, err := yaml.Marshal(values)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(store.Path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(store.Path, data, 0o600)
}
