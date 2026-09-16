package cli

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/manuel/miez-cli/internal/credentials"
	"github.com/manuel/miez-cli/internal/ghclient"
)

// newFixtureGitHubServer serves fixtureDir's contents as the tarball for
// every ref/commit request, standing in for a real GitHub repository now
// that the CLI no longer ships an embedded team.
func newFixtureGitHubServer(t *testing.T, fixtureDir, topLevelDir string) *httptest.Server {
	t.Helper()
	archive := buildFixtureTarGz(t, fixtureDir, topLevelDir)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case containsSegment(r.URL.Path, "commits"):
			_ = json.NewEncoder(w).Encode(map[string]string{"sha": "fixturesha"})
		case containsSegment(r.URL.Path, "tarball"):
			w.Write(archive)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)
	return server
}

func containsSegment(path, segment string) bool {
	for _, part := range splitSlash(path) {
		if part == segment {
			return true
		}
	}
	return false
}

func splitSlash(path string) []string {
	parts := []string{}
	current := ""
	for _, r := range path {
		if r == '/' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
			continue
		}
		current += string(r)
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

func buildFixtureTarGz(t *testing.T, fixtureDir, topLevelDir string) []byte {
	t.Helper()
	var buffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&buffer)
	tarWriter := tar.NewWriter(gzipWriter)
	err := filepath.WalkDir(fixtureDir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(fixtureDir, path)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		name := filepath.ToSlash(filepath.Join(topLevelDir, relative))
		if err := tarWriter.WriteHeader(&tar.Header{Name: name, Mode: 0o644, Size: int64(len(data))}); err != nil {
			return err
		}
		_, err = tarWriter.Write(data)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := tarWriter.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gzipWriter.Close(); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

// useFixtureGitHub wires app's team service to a fake GitHub server serving
// fixtureDir, for any owner/repo it is asked to download.
func useFixtureGitHub(t *testing.T, app *App, fixtureDir, topLevelDir string) {
	t.Helper()
	server := newFixtureGitHubServer(t, fixtureDir, topLevelDir)
	app.teams.GH = &ghclient.Client{APIBase: server.URL, HTTP: server.Client()}
	app.teams.Credentials = &credentials.Resolver{LookupEnv: func(string) (string, bool) { return "", false }}
}

// installFixtureTeam runs `team install` against the testdata/spec-driven-team
// fixture through a fake GitHub server.
func installFixtureTeam(t *testing.T, app *App) error {
	t.Helper()
	useFixtureGitHub(t, app, "testdata/spec-driven-team", "spec-driven-team")
	return app.Execute(context.Background(), []string{"team", "install", "https://github.com/example/spec-driven-team", "--targets", "copilot"})
}
