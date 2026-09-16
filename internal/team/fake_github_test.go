package team

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/manuel/miez-cli/internal/credentials"
	"github.com/manuel/miez-cli/internal/ghclient"
)

// fakeGitHub is an in-memory GitHub API + tarball server for tests. Repos
// are keyed by "owner/repo". requireToken, when set, makes every request
// require an exact "Bearer <token>" Authorization header.
type fakeGitHub struct {
	mu           sync.Mutex
	commits      map[string]map[string]string // repo -> ref -> sha
	archives     map[string]map[string][]byte // repo -> sha -> tar.gz bytes
	requireToken string
	server       *httptest.Server
}

func newFakeGitHub(t *testing.T) *fakeGitHub {
	t.Helper()
	fake := &fakeGitHub{
		commits:  map[string]map[string]string{},
		archives: map[string]map[string][]byte{},
	}
	fake.server = httptest.NewServer(http.HandlerFunc(fake.handle))
	t.Cleanup(fake.server.Close)
	return fake
}

func (fake *fakeGitHub) handle(w http.ResponseWriter, r *http.Request) {
	if fake.requireToken != "" {
		want := "Bearer " + fake.requireToken
		if r.Header.Get("Authorization") != want {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	}
	var owner, repo, kind, selector string
	parts := splitPath(r.URL.Path)
	if len(parts) < 4 || parts[0] != "repos" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	owner, repo, kind, selector = parts[1], parts[2], parts[3], ""
	if len(parts) > 4 {
		selector = parts[4]
	}
	key := owner + "/" + repo

	fake.mu.Lock()
	defer fake.mu.Unlock()
	switch kind {
	case "commits":
		sha, ok := fake.commits[key][selector]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"sha": sha})
	case "tarball":
		data, ok := fake.archives[key][selector]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Write(data)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func splitPath(path string) []string {
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

// setTeam registers ref -> commit and commit -> tarball(team) for owner/repo.
func (fake *fakeGitHub) setTeam(owner, repo, ref, commit string, files map[string]string) {
	fake.mu.Lock()
	defer fake.mu.Unlock()
	key := owner + "/" + repo
	if fake.commits[key] == nil {
		fake.commits[key] = map[string]string{}
	}
	if fake.archives[key] == nil {
		fake.archives[key] = map[string][]byte{}
	}
	fake.commits[key][ref] = commit
	fake.archives[key][commit] = buildTarGz(files)
}

func buildTarGz(files map[string]string) []byte {
	var buffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&buffer)
	tarWriter := tar.NewWriter(gzipWriter)
	for name, content := range files {
		header := &tar.Header{Name: name, Mode: 0o644, Size: int64(len(content))}
		_ = tarWriter.WriteHeader(header)
		_, _ = tarWriter.Write([]byte(content))
	}
	_ = tarWriter.Close()
	_ = gzipWriter.Close()
	return buffer.Bytes()
}

// newTestService builds a Service wired to fake's server and a credentials
// resolver backed by env, an in-memory map instead of the real environment.
func newTestService(root string, fake *fakeGitHub, env map[string]string) *Service {
	return &Service{
		Root: root,
		GH:   &ghclient.Client{APIBase: fake.server.URL, HTTP: fake.server.Client()},
		Credentials: &credentials.Resolver{
			LookupEnv: func(name string) (string, bool) {
				value, ok := env[name]
				return value, ok
			},
		},
	}
}

func demoTeamFiles(id, name string) map[string]string {
	return map[string]string{
		id + "/miez.yaml":            fmt.Sprintf("id: %s\nversion: 1.0.0\nname: %s\nauthor: test\n", id, name),
		id + "/miez.generated.yaml":  fmt.Sprintf("id: %s\nversion: 1.0.0\nname: %s\nauthor: test\nworkers:\n  - id: builder\n    kind: command\n    path: workers/builder.md\nworkflows:\n  - id: default\n    name: Default\n    path: workflows/default.md\n    phases:\n      - id: build\n        workers: [builder]\n", id, name),
		id + "/workers/builder.md":   "---\nid: builder\nkind: command\n---\n# Builder\n",
		id + "/workflows/default.md": "---\nid: default\nname: Default\nphases:\n  - id: build\n    workers: [builder]\n---\n# Default\n",
	}
}
