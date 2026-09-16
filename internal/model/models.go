package model

import (
	"fmt"
	"regexp"
	"strings"
)

// ModelOption is one canonical model choice miez knows how to render.
// The generated team index references a model by its stable ID here, never a provider's
// own name, so a team's manifest keeps working if Copilot's own naming
// changes.
type ModelOption struct {
	// ID is the stable identifier used in the generated index's model/default_model
	// fields. It is also what model listing and shell completion offer;
	// there is deliberately no separate display name to keep in sync.
	ID string `yaml:"id"`
	// Copilot is the value written into a Copilot agent's model
	// frontmatter, e.g. "Claude Sonnet 4.5 (copilot)".
	Copilot string `yaml:"copilot"`
}

// DefaultModelID is used when neither a worker nor its team sets a model.
const DefaultModelID = "gpt-5"

var modelIDPattern = regexp.MustCompile(`^[a-z0-9]+(?:[.-][a-z0-9]+)*$`)

// IsValidModelID reports whether value is an acceptable model id: lowercase
// letters, digits, dots, and hyphens. Unlike the strict kebab-case used for
// team/worker/skill/workflow ids, dots are allowed because provider model
// names commonly embed a version number (e.g. "gpt-5.1", "claude-sonnet-4.5").
func IsValidModelID(value string) bool {
	return modelIDPattern.MatchString(value)
}

// BuiltinModels are the model choices miez ships out of the box. Providers
// add, rename, and retire models often; a team is not limited to this list —
// it can add new models or correct a mapping in its own miez.yaml (see
// Team.Models) the moment a provider ships them, without waiting for a new
// miez release. This list only needs updating for models every team should
// get for free.
var BuiltinModels = []ModelOption{
	{ID: "gpt-5", Copilot: "GPT-5 (copilot)"},
	{ID: "gpt-5.1", Copilot: "GPT-5.1 (copilot)"},
	{ID: "claude-sonnet-4.5", Copilot: "Claude Sonnet 4.5 (copilot)"},
	{ID: "claude-opus-4.5", Copilot: "Claude Opus 4.5 (copilot)"},
	{ID: "gemini-3-pro", Copilot: "Gemini 3 Pro (copilot)"},
}

// MergedModels returns team's usable model catalog: BuiltinModels plus
// team.Models, in that order. A team.Models entry with the same id as a
// built-in one replaces it, so a team can also fix a wrong or outdated
// built-in mapping locally instead of waiting for a new miez release.
func MergedModels(team Team) []ModelOption {
	merged := make(map[string]ModelOption, len(BuiltinModels)+len(team.Models))
	order := make([]string, 0, len(BuiltinModels)+len(team.Models))
	for _, option := range BuiltinModels {
		merged[option.ID] = option
		order = append(order, option.ID)
	}
	for _, option := range team.Models {
		if _, exists := merged[option.ID]; !exists {
			order = append(order, option.ID)
		}
		merged[option.ID] = option
	}
	resolved := make([]ModelOption, len(order))
	for index, id := range order {
		resolved[index] = merged[id]
	}
	return resolved
}

// FindModel looks up a model by its stable id in team's merged catalog.
func FindModel(team Team, id string) (ModelOption, bool) {
	for _, option := range MergedModels(team) {
		if option.ID == id {
			return option, true
		}
	}
	return ModelOption{}, false
}

// ModelIDs returns every model id in team's merged catalog, in listing order.
func ModelIDs(team Team) []string {
	merged := MergedModels(team)
	ids := make([]string, len(merged))
	for index, option := range merged {
		ids[index] = option.ID
	}
	return ids
}

// NameFor returns option's Copilot model name, or an error if option has
// none set.
func (option ModelOption) NameFor() (string, error) {
	if strings.TrimSpace(option.Copilot) == "" {
		return "", fmt.Errorf("model %q has no copilot equivalent", option.ID)
	}
	return option.Copilot, nil
}
