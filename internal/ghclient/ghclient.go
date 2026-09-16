// Package ghclient talks to the GitHub REST API to resolve a ref to a
// commit and download a repository tarball, carrying an optional bearer
// credential across the api.github.com -> codeload.github.com redirect.
package ghclient

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	requestTimeout            = 30 * time.Second
	maxCompressedArchiveBytes = 64 << 20
	maxExpandedArchiveBytes   = 256 << 20
	maxArchiveFileBytes       = 32 << 20
	maxArchiveFiles           = 10_000
	DefaultAPIBase            = "https://api.github.com"
)

// Client downloads GitHub repository content, optionally authenticated.
type Client struct {
	// APIBase overrides the GitHub API base URL; tests point this at a
	// local httptest.Server. Empty means DefaultAPIBase.
	APIBase string
	HTTP    *http.Client
}

// New creates a Client with a redirect policy that reattaches the
// Authorization header only when a redirect stays within *.github.com
// hosts (Go's default client strips it on any host change).
func New() *Client {
	client := &http.Client{Timeout: requestTimeout}
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) == 0 {
			return nil
		}
		if len(via) > 10 {
			return errors.New("too many redirects")
		}
		if isGitHubHost(via[0].URL.Hostname()) && isGitHubHost(req.URL.Hostname()) {
			if auth := via[0].Header.Get("Authorization"); auth != "" {
				req.Header.Set("Authorization", auth)
			}
		}
		return nil
	}
	return &Client{HTTP: client}
}

func isGitHubHost(host string) bool {
	return host == "github.com" || host == "api.github.com" || host == "codeload.github.com" ||
		strings.HasSuffix(host, ".github.com")
}

func (client *Client) apiBase() string {
	if client.APIBase != "" {
		return client.APIBase
	}
	return DefaultAPIBase
}

func (client *Client) httpClient() *http.Client {
	if client.HTTP != nil {
		return client.HTTP
	}
	return http.DefaultClient
}

// ResolveCommit resolves ref (branch, tag, or sha) to its current commit sha.
func (client *Client) ResolveCommit(ctx context.Context, owner, repository, ref, token string) (string, error) {
	requestURL := fmt.Sprintf("%s/repos/%s/%s/commits/%s", client.apiBase(), owner, repository, url.PathEscape(ref))
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return "", fmt.Errorf("create commit request: %w", err)
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	response, err := client.httpClient().Do(request)
	if err != nil {
		return "", fmt.Errorf("resolve commit for %s: %w", ref, err)
	}
	defer response.Body.Close()
	if err := checkStatus(response, "resolve commit"); err != nil {
		return "", err
	}
	var payload struct {
		SHA string `json:"sha"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return "", fmt.Errorf("parse commit response: %w", err)
	}
	if payload.SHA == "" {
		return "", fmt.Errorf("GitHub returned no commit sha for ref %q", ref)
	}
	return payload.SHA, nil
}

// DownloadTarball downloads the tarball for commit and extracts it below
// destination.
func (client *Client) DownloadTarball(ctx context.Context, owner, repository, commit, token, destination string) error {
	requestURL := fmt.Sprintf("%s/repos/%s/%s/tarball/%s", client.apiBase(), owner, repository, url.PathEscape(commit))
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return fmt.Errorf("create tarball request: %w", err)
	}
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	response, err := client.httpClient().Do(request)
	if err != nil {
		return fmt.Errorf("download tarball: %w", err)
	}
	defer response.Body.Close()
	if err := checkStatus(response, "download tarball"); err != nil {
		return err
	}

	compressed := &countingLimitReader{Reader: response.Body, Limit: maxCompressedArchiveBytes + 1}
	if err := extractArchive(compressed, destination); err != nil {
		return err
	}
	if compressed.Count > maxCompressedArchiveBytes {
		return fmt.Errorf("GitHub archive exceeds %d byte compressed limit", maxCompressedArchiveBytes)
	}
	return nil
}

func checkStatus(response *http.Response, action string) error {
	if response.StatusCode == http.StatusOK {
		return nil
	}
	if response.StatusCode == http.StatusUnauthorized || response.StatusCode == http.StatusForbidden {
		return fmt.Errorf("%s: authentication failed (HTTP %s); check the resolved GitHub credential", action, response.Status)
	}
	if response.StatusCode == http.StatusNotFound {
		return fmt.Errorf("%s: not found (HTTP %s); the repository, ref, or path may not exist, or requires a credential", action, response.Status)
	}
	if response.StatusCode == http.StatusTooManyRequests {
		return fmt.Errorf("%s: GitHub rate limit exceeded", action)
	}
	return fmt.Errorf("%s: HTTP %s", action, response.Status)
}

func extractArchive(source io.Reader, destination string) error {
	compressed, err := gzip.NewReader(source)
	if err != nil {
		return fmt.Errorf("read GitHub archive: %w", err)
	}
	defer compressed.Close()
	reader := tar.NewReader(compressed)
	fileCount := 0
	var expandedBytes int64
	for {
		header, err := reader.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("read GitHub archive entry: %w", err)
		}
		if header.Typeflag == tar.TypeXGlobalHeader {
			continue
		}
		relative, err := safeArchivePath(header.Name)
		if err != nil {
			return err
		}
		path := filepath.Join(destination, relative)
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(path, 0o755); err != nil {
				return fmt.Errorf("create archive directory: %w", err)
			}
		case tar.TypeReg, tar.TypeRegA:
			if header.Size < 0 || header.Size > maxArchiveFileBytes {
				return fmt.Errorf("GitHub archive file %q exceeds %d byte limit", header.Name, maxArchiveFileBytes)
			}
			if expandedBytes+header.Size > maxExpandedArchiveBytes {
				return fmt.Errorf("GitHub archive exceeds %d byte expanded limit", maxExpandedArchiveBytes)
			}
			fileCount++
			if fileCount > maxArchiveFiles {
				return fmt.Errorf("GitHub archive exceeds %d file limit", maxArchiveFiles)
			}
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				return fmt.Errorf("create archive parent: %w", err)
			}
			file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
			if err != nil {
				return fmt.Errorf("create archive file: %w", err)
			}
			copied, copyErr := io.CopyN(file, reader, header.Size)
			closeErr := file.Close()
			if copyErr != nil {
				return fmt.Errorf("extract archive file: %w", copyErr)
			}
			if closeErr != nil {
				return fmt.Errorf("close archive file: %w", closeErr)
			}
			if copied != header.Size {
				return fmt.Errorf("extract archive file %q: short file", header.Name)
			}
			expandedBytes += copied
		default:
			return fmt.Errorf("GitHub archive contains unsupported entry %q", header.Name)
		}
	}
}

type countingLimitReader struct {
	Reader io.Reader
	Limit  int64
	Count  int64
}

func (reader *countingLimitReader) Read(buffer []byte) (int, error) {
	if reader.Count >= reader.Limit {
		return 0, io.ErrShortBuffer
	}
	remaining := reader.Limit - reader.Count
	if int64(len(buffer)) > remaining {
		buffer = buffer[:remaining]
	}
	count, err := reader.Reader.Read(buffer)
	reader.Count += int64(count)
	return count, err
}

func safeArchivePath(value string) (string, error) {
	clean := filepath.Clean(filepath.FromSlash(value))
	if !filepath.IsLocal(clean) || clean == "." {
		return "", fmt.Errorf("GitHub archive path %q escapes staging directory", value)
	}
	return clean, nil
}

// SingleTopLevelDir returns the archive's single top-level directory.
// GitHub's tarballs always extract into exactly one such directory.
func SingleTopLevelDir(root string) (string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return "", fmt.Errorf("inspect downloaded archive: %w", err)
	}
	var dirs []string
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
		}
	}
	if len(dirs) != 1 {
		return "", fmt.Errorf("GitHub archive must contain exactly one top-level directory, found %d", len(dirs))
	}
	return filepath.Join(root, dirs[0]), nil
}
