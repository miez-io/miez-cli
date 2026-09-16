// Package credentials resolves a GitHub credential for module access,
// mirroring APM's own GitHub-module-access chain but renamed so a repo with
// both `apm` and `miez` installed never has one silently authenticate the
// other.
package credentials

import (
	"context"
	"os/exec"
	"strings"
)

// Credential is a resolved GitHub token and the chain tier that produced it.
type Credential struct {
	Token  string
	Source string
}

// Chain tier names, in priority order, reported under a verbose flag.
const (
	SourcePerOrg   = "GITHUB_MIEZ_PAT_<ORG>"
	SourceMiezPAT  = "GITHUB_MIEZ_PAT"
	SourceGitHub   = "GITHUB_TOKEN"
	SourceGH       = "GH_TOKEN"
	SourceGhCLI    = "gh auth token"
	orgEnvPrefix   = "GITHUB_MIEZ_PAT_"
	perOrgEnvChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
)

// Resolver resolves a GitHub credential. LookupEnv and GhAuthToken are
// injectable so tests never depend on the real process environment or the
// gh CLI.
type Resolver struct {
	LookupEnv   func(string) (string, bool)
	GhAuthToken func(ctx context.Context, hostname string) (string, error)
}

// NewResolver creates a Resolver backed by the real process environment and
// the `gh` CLI.
func NewResolver() *Resolver {
	return &Resolver{
		LookupEnv:   osLookupEnv,
		GhAuthToken: ghCLIAuthToken,
	}
}

// Resolve walks the chain for org (empty when the target org is not known)
// against hostname. ok is false, with no error, when no tier resolves a
// credential -- that is not itself a failure, since public repositories need
// none. No git-credential-helper fallback is ever attempted.
func (resolver *Resolver) Resolve(ctx context.Context, org, hostname string) (Credential, bool, error) {
	if org != "" {
		if token, ok := resolver.lookup(orgEnvPrefix + normalizeOrgEnv(org)); ok {
			return Credential{Token: token, Source: SourcePerOrg}, true, nil
		}
	}
	if token, ok := resolver.lookup("GITHUB_MIEZ_PAT"); ok {
		return Credential{Token: token, Source: SourceMiezPAT}, true, nil
	}
	if token, ok := resolver.lookup("GITHUB_TOKEN"); ok {
		return Credential{Token: token, Source: SourceGitHub}, true, nil
	}
	if token, ok := resolver.lookup("GH_TOKEN"); ok {
		return Credential{Token: token, Source: SourceGH}, true, nil
	}
	if resolver.GhAuthToken != nil {
		token, err := resolver.GhAuthToken(ctx, hostname)
		if err == nil && strings.TrimSpace(token) != "" {
			return Credential{Token: strings.TrimSpace(token), Source: SourceGhCLI}, true, nil
		}
	}
	return Credential{}, false, nil
}

func (resolver *Resolver) lookup(name string) (string, bool) {
	if resolver.LookupEnv == nil {
		return "", false
	}
	value, ok := resolver.LookupEnv(name)
	if !ok || strings.TrimSpace(value) == "" {
		return "", false
	}
	return value, true
}

// normalizeOrgEnv upper-cases org and replaces any character not valid in an
// environment variable name with an underscore.
func normalizeOrgEnv(org string) string {
	var builder strings.Builder
	for _, character := range strings.ToUpper(org) {
		if strings.ContainsRune(perOrgEnvChars, character) {
			builder.WriteRune(character)
		} else {
			builder.WriteRune('_')
		}
	}
	return builder.String()
}

func ghCLIAuthToken(ctx context.Context, hostname string) (string, error) {
	args := []string{"auth", "token"}
	if hostname != "" {
		args = append(args, "--hostname", hostname)
	}
	command := exec.CommandContext(ctx, "gh", args...)
	output, err := command.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}
