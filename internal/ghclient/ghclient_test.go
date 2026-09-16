package ghclient_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/manuel/miez-cli/internal/ghclient"
)

func TestResolveCommitSendsBearerTokenAndParsesSHA(t *testing.T) {
	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/repos/owner/repo/commits/main" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"sha": "deadbeef"})
	}))
	defer server.Close()

	client := &ghclient.Client{APIBase: server.URL, HTTP: server.Client()}
	sha, err := client.ResolveCommit(context.Background(), "owner", "repo", "main", "secret-token")
	if err != nil {
		t.Fatal(err)
	}
	if sha != "deadbeef" {
		t.Fatalf("sha = %q, want deadbeef", sha)
	}
	if gotAuth != "Bearer secret-token" {
		t.Fatalf("Authorization = %q, want Bearer secret-token", gotAuth)
	}
}

func TestResolveCommitReportsAuthFailureClearly(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client := &ghclient.Client{APIBase: server.URL, HTTP: server.Client()}
	if _, err := client.ResolveCommit(context.Background(), "owner", "repo", "main", ""); err == nil {
		t.Fatal("ResolveCommit succeeded despite a 401 response")
	}
}

func TestDownloadTarballExtractsArchive(t *testing.T) {
	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Write(buildTarGz(t, map[string]string{
			"owner-repo-deadbeef/miez.generated.yaml": "id: spec-driven-team\nversion: 1.0.0\nname: Spec Driven\n",
			"owner-repo-deadbeef/workers/starter.md":  "---\n---\n# Starter\n",
		}))
	}))
	defer server.Close()

	client := &ghclient.Client{APIBase: server.URL, HTTP: server.Client()}
	destination := t.TempDir()
	if err := client.DownloadTarball(context.Background(), "owner", "repo", "deadbeef", "secret-token", destination); err != nil {
		t.Fatal(err)
	}
	if gotAuth != "Bearer secret-token" {
		t.Fatalf("Authorization = %q, want Bearer secret-token", gotAuth)
	}
	top, err := ghclient.SingleTopLevelDir(destination)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(top, "miez.generated.yaml")); err != nil {
		t.Fatalf("generated team index missing after extraction: %v", err)
	}
}

func TestDownloadTarballRejectsPathEscape(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(buildTarGz(t, map[string]string{"../outside.md": "oops"}))
	}))
	defer server.Close()

	client := &ghclient.Client{APIBase: server.URL, HTTP: server.Client()}
	if err := client.DownloadTarball(context.Background(), "owner", "repo", "deadbeef", "", t.TempDir()); err == nil {
		t.Fatal("DownloadTarball accepted a path-escaping archive entry")
	}
}

func TestNewReattachesAuthorizationAcrossGitHubHostRedirect(t *testing.T) {
	client := ghclient.New()
	from, _ := url.Parse("https://api.github.com/repos/o/r/tarball/main")
	to, _ := url.Parse("https://codeload.github.com/o/r/tar.gz/main")
	initial := &http.Request{URL: from, Header: http.Header{"Authorization": []string{"Bearer secret"}}}
	next := &http.Request{URL: to, Header: http.Header{}}
	if err := client.HTTP.CheckRedirect(next, []*http.Request{initial}); err != nil {
		t.Fatal(err)
	}
	if next.Header.Get("Authorization") != "Bearer secret" {
		t.Fatalf("Authorization not reattached across GitHub-host redirect")
	}
}

func TestNewDropsAuthorizationAcrossNonGitHubHostRedirect(t *testing.T) {
	client := ghclient.New()
	from, _ := url.Parse("https://api.github.com/repos/o/r/tarball/main")
	to, _ := url.Parse("https://evil.example.com/steal")
	initial := &http.Request{URL: from, Header: http.Header{"Authorization": []string{"Bearer secret"}}}
	next := &http.Request{URL: to, Header: http.Header{}}
	if err := client.HTTP.CheckRedirect(next, []*http.Request{initial}); err != nil {
		t.Fatal(err)
	}
	if next.Header.Get("Authorization") != "" {
		t.Fatal("Authorization leaked to a non-GitHub redirect target")
	}
}

func buildTarGz(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&buffer)
	tarWriter := tar.NewWriter(gzipWriter)
	for name, content := range files {
		header := &tar.Header{Name: name, Mode: 0o644, Size: int64(len(content))}
		if err := tarWriter.WriteHeader(header); err != nil {
			t.Fatal(err)
		}
		if _, err := tarWriter.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := tarWriter.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gzipWriter.Close(); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}
