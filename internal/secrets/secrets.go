// Package secrets resolves MCP tool env placeholders declared in the team package/index.
// against an override map or the process environment, prompting
// interactively for a value that resolves from neither, and persisting
// supplied values only to an untracked, workspace-local file.
package secrets

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/manuel/miez-cli/internal/model"
)

var placeholderPatterns = []*regexp.Regexp{
	regexp.MustCompile(`^<([A-Za-z_][A-Za-z0-9_]*)>$`),
	regexp.MustCompile(`^\$\{env:([A-Za-z_][A-Za-z0-9_]*)\}$`),
	regexp.MustCompile(`^\$\{([A-Za-z_][A-Za-z0-9_]*)\}$`),
}

// ParsePlaceholder extracts the source variable name from one of the three
// accepted placeholder syntaxes: "<VAR>", "${VAR}", "${env:VAR}". ok is
// false when value does not match any of them.
func ParsePlaceholder(value string) (variable string, ok bool) {
	trimmed := strings.TrimSpace(value)
	for _, pattern := range placeholderPatterns {
		if match := pattern.FindStringSubmatch(trimmed); match != nil {
			return match[1], true
		}
	}
	return "", false
}

// IsMasked reports whether a variable name should be masked when prompted
// for interactively: its name contains "token" or "key", case-insensitive.
func IsMasked(variable string) bool {
	lower := strings.ToLower(variable)
	return strings.Contains(lower, "token") || strings.Contains(lower, "key")
}

// Resolver resolves a source variable name against an override map, then
// the process environment.
type Resolver struct {
	Overrides map[string]string
	LookupEnv func(string) (string, bool)
}

// Resolve looks up variable in Overrides, then LookupEnv.
func (resolver Resolver) Resolve(variable string) (string, bool) {
	if value, ok := resolver.Overrides[variable]; ok && value != "" {
		return value, true
	}
	if resolver.LookupEnv != nil {
		if value, ok := resolver.LookupEnv(variable); ok && value != "" {
			return value, true
		}
	}
	return "", false
}

// Requirement is one MCP server's env key and the source variable its
// placeholder names.
type Requirement struct {
	ServerID string
	EnvKey   string
	Variable string
}

// RequiredFor collects every env placeholder requirement for the given mcp
// ids (a worker's Tools field), sorted for deterministic prompting order.
func RequiredFor(team model.Team, mcpIDs []string) ([]Requirement, error) {
	servers := map[string]model.MCPServer{}
	for _, server := range team.MCP {
		servers[server.ID] = server
	}
	requirements := []Requirement{}
	seen := map[string]struct{}{}
	for _, id := range mcpIDs {
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		server, ok := servers[id]
		if !ok {
			return nil, fmt.Errorf("unknown mcp id %q", id)
		}
		keys := make([]string, 0, len(server.Env))
		for key := range server.Env {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			variable, ok := ParsePlaceholder(server.Env[key])
			if !ok {
				return nil, fmt.Errorf("mcp %q env %q: %q is not a valid placeholder (use <VAR>, ${VAR}, or ${env:VAR})", id, key, server.Env[key])
			}
			requirements = append(requirements, Requirement{ServerID: id, EnvKey: key, Variable: variable})
		}
	}
	return requirements, nil
}

// Prompt reads one interactively supplied value for variable. masked
// indicates whether input should be hidden.
type Prompt func(variable string, masked bool) (string, error)

// Resolve resolves every requirement's source variable against resolver,
// falling back to prompt when interactive is true and a value resolves
// from neither the overrides nor the environment. It returns the resolved
// value per source variable name (not per requirement, since two
// requirements may share one source variable).
//
// When interactive is false, any unresolved variable fails the whole
// resolution with a clear error naming every missing value, instead of
// silently proceeding with an unresolved placeholder or hanging.
func Resolve(requirements []Requirement, resolver Resolver, interactive bool, prompt Prompt) (map[string]string, error) {
	resolved := map[string]string{}
	missing := []string{}
	for _, requirement := range requirements {
		if _, done := resolved[requirement.Variable]; done {
			continue
		}
		if value, ok := resolver.Resolve(requirement.Variable); ok {
			resolved[requirement.Variable] = value
			continue
		}
		if !interactive {
			missing = append(missing, requirement.Variable)
			continue
		}
		value, err := prompt(requirement.Variable, IsMasked(requirement.Variable))
		if err != nil {
			return nil, err
		}
		resolved[requirement.Variable] = value
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return nil, fmt.Errorf("missing required MCP configuration value(s): %s (set as an environment variable, or supply them interactively)", strings.Join(missing, ", "))
	}
	return resolved, nil
}
