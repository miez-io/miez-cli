package credentials_test

import (
	"context"
	"testing"

	"github.com/manuel/miez-cli/internal/credentials"
)

func envLookup(values map[string]string) func(string) (string, bool) {
	return func(name string) (string, bool) {
		value, ok := values[name]
		return value, ok
	}
}

func TestResolvePrefersPerOrgOverEverythingElse(t *testing.T) {
	resolver := &credentials.Resolver{
		LookupEnv: envLookup(map[string]string{
			"GITHUB_MIEZ_PAT_ACME": "org-token",
			"GITHUB_MIEZ_PAT":      "generic-miez-token",
			"GITHUB_TOKEN":         "generic-token",
			"GH_TOKEN":             "gh-env-token",
		}),
	}
	credential, ok, err := resolver.Resolve(context.Background(), "acme", "github.com")
	if err != nil || !ok {
		t.Fatalf("Resolve() = (%v, %v, %v)", credential, ok, err)
	}
	if credential.Token != "org-token" || credential.Source != credentials.SourcePerOrg {
		t.Fatalf("credential = %#v, want org-token via per-org tier", credential)
	}
}

func TestResolveSkipsPerOrgTierWhenOrgUnknown(t *testing.T) {
	resolver := &credentials.Resolver{
		LookupEnv: envLookup(map[string]string{
			"GITHUB_MIEZ_PAT_ACME": "org-token",
			"GITHUB_MIEZ_PAT":      "generic-miez-token",
		}),
	}
	credential, ok, err := resolver.Resolve(context.Background(), "", "github.com")
	if err != nil || !ok {
		t.Fatalf("Resolve() = (%v, %v, %v)", credential, ok, err)
	}
	if credential.Source != credentials.SourceMiezPAT {
		t.Fatalf("source = %q, want %q", credential.Source, credentials.SourceMiezPAT)
	}
}

func TestResolveFallsThroughTiersInOrder(t *testing.T) {
	cases := []struct {
		name   string
		values map[string]string
		source string
	}{
		{"miez pat", map[string]string{"GITHUB_MIEZ_PAT": "a"}, credentials.SourceMiezPAT},
		{"github token", map[string]string{"GITHUB_TOKEN": "a"}, credentials.SourceGitHub},
		{"gh token env", map[string]string{"GH_TOKEN": "a"}, credentials.SourceGH},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			resolver := &credentials.Resolver{LookupEnv: envLookup(testCase.values)}
			credential, ok, err := resolver.Resolve(context.Background(), "acme", "github.com")
			if err != nil || !ok {
				t.Fatalf("Resolve() = (%v, %v, %v)", credential, ok, err)
			}
			if credential.Source != testCase.source {
				t.Fatalf("source = %q, want %q", credential.Source, testCase.source)
			}
		})
	}
}

func TestResolveFallsBackToGhCLI(t *testing.T) {
	resolver := &credentials.Resolver{
		LookupEnv: envLookup(map[string]string{}),
		GhAuthToken: func(ctx context.Context, hostname string) (string, error) {
			if hostname != "github.com" {
				t.Fatalf("hostname = %q, want github.com", hostname)
			}
			return "cli-token\n", nil
		},
	}
	credential, ok, err := resolver.Resolve(context.Background(), "acme", "github.com")
	if err != nil || !ok {
		t.Fatalf("Resolve() = (%v, %v, %v)", credential, ok, err)
	}
	if credential.Token != "cli-token" || credential.Source != credentials.SourceGhCLI {
		t.Fatalf("credential = %#v", credential)
	}
}

func TestResolveReturnsNotOkWhenNothingResolves(t *testing.T) {
	resolver := &credentials.Resolver{
		LookupEnv: envLookup(map[string]string{}),
		GhAuthToken: func(ctx context.Context, hostname string) (string, error) {
			return "", context.DeadlineExceeded
		},
	}
	credential, ok, err := resolver.Resolve(context.Background(), "acme", "github.com")
	if err != nil {
		t.Fatalf("Resolve() error = %v, want nil", err)
	}
	if ok {
		t.Fatalf("Resolve() ok = true, credential = %#v, want false", credential)
	}
}
