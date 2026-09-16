// Package lockfile owns the committed .miez/miez.lock.yaml: per installed team,
// the resolved repo URL and commit, package version, deployed file list, and a
// content hash per deployed file. It is what lets update, audit, and outdated
// compare against a known-good install without re-resolving any remote ref.
package lockfile

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/manuel/miez-cli/internal/artifacts"
	"gopkg.in/yaml.v3"
)

// FileName is the lockfile's fixed name below the workspace's .miez directory.
const FileName = "miez.lock.yaml"

// Entry is one installed team's reproducible install record.
type Entry struct {
	RepoURL       string            `yaml:"repo_url"`
	Ref           string            `yaml:"ref"`
	Commit        string            `yaml:"commit"`
	Version       string            `yaml:"version,omitempty"`
	DeployedFiles []string          `yaml:"deployed_files"`
	FileHashes    map[string]string `yaml:"file_hashes"`
}

// Lockfile is the committed reproducible-install record.
type Lockfile struct {
	Teams map[string]Entry `yaml:"teams,omitempty"`
}

// Path returns the lockfile path below the workspace's .miez directory.
func Path(workspaceRoot string) string {
	return filepath.Join(workspaceRoot, ".miez", FileName)
}

// Load reads the lockfile, returning an empty Lockfile when absent.
func Load(workspaceRoot string) (Lockfile, error) {
	data, err := os.ReadFile(Path(workspaceRoot))
	if errors.Is(err, os.ErrNotExist) {
		return Lockfile{Teams: map[string]Entry{}}, nil
	}
	if err != nil {
		return Lockfile{}, fmt.Errorf("read %s: %w", FileName, err)
	}
	var lock Lockfile
	if err := yaml.Unmarshal(data, &lock); err != nil {
		return Lockfile{}, fmt.Errorf("parse %s: %w", FileName, err)
	}
	if lock.Teams == nil {
		lock.Teams = map[string]Entry{}
	}
	return lock, nil
}

// Save writes the lockfile atomically.
func Save(workspaceRoot string, lock Lockfile) error {
	data, err := yaml.Marshal(lock)
	if err != nil {
		return fmt.Errorf("encode %s: %w", FileName, err)
	}
	path := Path(workspaceRoot)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create lockfile directory: %w", err)
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".miez-lock-")
	if err != nil {
		return fmt.Errorf("create temporary %s: %w", FileName, err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o644); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("set %s permissions: %w", FileName, err)
	}
	if _, err := temporary.Write(data); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("write %s: %w", FileName, err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close temporary %s: %w", FileName, err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("replace %s: %w", FileName, err)
	}
	return nil
}

// HashFile returns the hex sha256 digest of a deployed file's contents.
func HashFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("hash %s: %w", path, err)
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

// BuildEntry walks teamRoot's deployed files and hashes every one of them,
// producing the entry Use/Update write to the lockfile.
func BuildEntry(teamRoot, repoURL, ref, commit, version string) (Entry, error) {
	files, err := artifacts.WalkFiles(teamRoot)
	if err != nil {
		return Entry{}, err
	}
	hashes := make(map[string]string, len(files))
	for _, relative := range files {
		hash, err := HashFile(filepath.Join(teamRoot, filepath.FromSlash(relative)))
		if err != nil {
			return Entry{}, err
		}
		hashes[relative] = hash
	}
	return Entry{
		RepoURL:       repoURL,
		Ref:           ref,
		Commit:        commit,
		Version:       version,
		DeployedFiles: files,
		FileHashes:    hashes,
	}, nil
}
