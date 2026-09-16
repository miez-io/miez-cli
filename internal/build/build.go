// Package build compiles a team authoring package into its installable index.
package build

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/manuel/miez-cli/internal/artifacts"
	"github.com/manuel/miez-cli/internal/manifest"
	"github.com/manuel/miez-cli/internal/model"
	"github.com/manuel/miez-cli/internal/validate"
	"gopkg.in/yaml.v3"
)

// Plan contains a validated team catalog and its deterministic YAML output.
type Plan struct {
	Team model.Team
	Data []byte
}

type workerFrontmatter struct {
	Kind        string   `yaml:"kind"`
	Name        string   `yaml:"name,omitempty"`
	Description string   `yaml:"description,omitempty"`
	Model       string   `yaml:"model,omitempty"`
	Skills      []string `yaml:"skills,omitempty"`
	Tools       []string `yaml:"tools,omitempty"`
}

type workflowFrontmatter struct {
	Name        string        `yaml:"name"`
	Description string        `yaml:"description,omitempty"`
	Phases      []model.Phase `yaml:"phases"`
}

// Build validates a team package and returns the generated team index bytes.
func Build(root string) (Plan, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return Plan{}, fmt.Errorf("resolve team package: %w", err)
	}

	packageManifest, err := manifest.LoadTeamPackage(root)
	if err != nil {
		return Plan{}, err
	}

	team := model.Team{
		ID:           packageManifest.ID,
		Version:      packageManifest.Version,
		Name:         packageManifest.Name,
		Description:  packageManifest.Description,
		Author:       packageManifest.Author,
		DefaultModel: packageManifest.DefaultModel,
		Models:       append([]model.ModelOption{}, packageManifest.Models...),
		MCP:          append([]model.MCPServer{}, packageManifest.MCP...),
		Skills:       []model.Skill{},
		Workers:      []model.Worker{},
		Workflows:    []model.Workflow{},
	}

	if err := validatePackageModels(team); err != nil {
		return Plan{}, err
	}

	team.Skills, err = loadSkills(root)
	if err != nil {
		return Plan{}, err
	}
	team.Workers, err = loadWorkers(root)
	if err != nil {
		return Plan{}, err
	}
	team.Workflows, err = loadWorkflows(root)
	if err != nil {
		return Plan{}, err
	}
	if len(team.Workflows) == 0 {
		return Plan{}, errors.New("team package must contain at least one workflow Markdown file")
	}

	if err := validate.Package(root, team); err != nil {
		return Plan{}, err
	}
	teamDir := artifacts.TeamDir{Root: root, Team: team}
	if err := validate.Team(teamDir).Error(); err != nil {
		return Plan{}, err
	}

	data, err := yaml.Marshal(team)
	if err != nil {
		return Plan{}, fmt.Errorf("encode generated %s: %w", artifacts.TeamIndexFileName, err)
	}
	data = append([]byte(artifacts.GeneratedTeamIndexHeader), data...)
	return Plan{Team: team, Data: data}, nil
}

// Apply writes a successful build plan atomically beside miez.yaml.
func Apply(root string, plan Plan) error {
	root, err := filepath.Abs(root)
	if err != nil {
		return fmt.Errorf("resolve team package: %w", err)
	}
	temporary, err := os.CreateTemp(root, ".miez-team-build-*")
	if err != nil {
		return fmt.Errorf("create generated %s staging file: %w", artifacts.TeamIndexFileName, err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o644); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("set generated %s mode: %w", artifacts.TeamIndexFileName, err)
	}
	if _, err := temporary.Write(plan.Data); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("write generated %s: %w", artifacts.TeamIndexFileName, err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close generated %s: %w", artifacts.TeamIndexFileName, err)
	}
	if err := os.Rename(temporaryPath, filepath.Join(root, artifacts.TeamIndexFileName)); err != nil {
		return fmt.Errorf("replace generated %s: %w", artifacts.TeamIndexFileName, err)
	}
	return nil
}

func validatePackageModels(team model.Team) error {
	seen := map[string]struct{}{}
	for _, option := range team.Models {
		if !model.IsValidModelID(option.ID) {
			return fmt.Errorf("model id %q is not valid", option.ID)
		}
		if _, exists := seen[option.ID]; exists {
			return fmt.Errorf("duplicate model id %q", option.ID)
		}
		seen[option.ID] = struct{}{}
		if strings.TrimSpace(option.Copilot) == "" {
			return fmt.Errorf("model %q must set a copilot name", option.ID)
		}
	}
	if team.DefaultModel != "" {
		if _, ok := model.FindModel(team, team.DefaultModel); !ok {
			return fmt.Errorf("default_model %q is not supported", team.DefaultModel)
		}
	}
	return nil
}

func loadWorkers(root string) ([]model.Worker, error) {
	paths, err := markdownFiles(root, "workers")
	if err != nil {
		return nil, err
	}
	workers := make([]model.Worker, 0, len(paths))
	seen := map[string]struct{}{}
	for _, path := range paths {
		markdown, err := artifacts.ReadMarkdown(root, path)
		if err != nil {
			return nil, fmt.Errorf("worker %s: %w", path, err)
		}
		var frontmatter workerFrontmatter
		if err := decodeFrontmatter(markdown.Metadata, &frontmatter, true); err != nil {
			return nil, fmt.Errorf("worker %s: %w", path, err)
		}
		id := artifacts.IDFromPath(path)
		if _, exists := seen[id]; exists {
			return nil, fmt.Errorf("duplicate worker id %q", id)
		}
		seen[id] = struct{}{}
		workers = append(workers, model.Worker{
			ID:          id,
			Kind:        frontmatter.Kind,
			Path:        path,
			Description: frontmatter.Description,
			Model:       frontmatter.Model,
			Skills:      append([]string{}, frontmatter.Skills...),
			Tools:       append([]string{}, frontmatter.Tools...),
		})
	}
	sort.Slice(workers, func(left, right int) bool { return workers[left].ID < workers[right].ID })
	return workers, nil
}

func loadSkills(root string) ([]model.Skill, error) {
	paths, err := markdownFiles(root, "skills")
	if err != nil {
		return nil, err
	}
	skills := make([]model.Skill, 0, len(paths))
	seen := map[string]struct{}{}
	for _, path := range paths {
		if filepath.Base(filepath.FromSlash(path)) != "SKILL.md" {
			continue
		}
		parts := strings.Split(path, "/")
		if len(parts) != 3 || parts[0] != "skills" || parts[2] != "SKILL.md" {
			return nil, fmt.Errorf("skill %s must use skills/<skill-id>/SKILL.md", path)
		}
		if _, err := artifacts.ReadMarkdown(root, path); err != nil {
			return nil, fmt.Errorf("skill %s: %w", path, err)
		}
		id := artifacts.IDFromPath(path)
		if _, exists := seen[id]; exists {
			return nil, fmt.Errorf("duplicate skill id %q", id)
		}
		seen[id] = struct{}{}
		skills = append(skills, model.Skill{ID: id, Path: path})
	}
	sort.Slice(skills, func(left, right int) bool { return skills[left].ID < skills[right].ID })
	return skills, nil
}

func loadWorkflows(root string) ([]model.Workflow, error) {
	paths, err := markdownFiles(root, "workflows")
	if err != nil {
		return nil, err
	}
	workflows := make([]model.Workflow, 0, len(paths))
	seen := map[string]struct{}{}
	for _, path := range paths {
		markdown, err := artifacts.ReadMarkdown(root, path)
		if err != nil {
			return nil, fmt.Errorf("workflow %s: %w", path, err)
		}
		var frontmatter workflowFrontmatter
		if err := decodeFrontmatter(markdown.Metadata, &frontmatter, true); err != nil {
			return nil, fmt.Errorf("workflow %s: %w", path, err)
		}
		id := artifacts.IDFromPath(path)
		if _, exists := seen[id]; exists {
			return nil, fmt.Errorf("duplicate workflow id %q", id)
		}
		seen[id] = struct{}{}
		workflows = append(workflows, model.Workflow{
			ID:     id,
			Name:   frontmatter.Name,
			Path:   path,
			Phases: append([]model.Phase{}, frontmatter.Phases...),
		})
	}
	sort.Slice(workflows, func(left, right int) bool { return workflows[left].ID < workflows[right].ID })
	return workflows, nil
}

func markdownFiles(root, directory string) ([]string, error) {
	base := filepath.Join(root, directory)
	info, err := os.Lstat(base)
	if errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("required artifact directory %s is missing", directory)
	}
	if err != nil {
		return nil, fmt.Errorf("inspect artifact directory %s: %w", directory, err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("artifact directory %s must not be a symlink", directory)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("artifact path %s is not a directory", directory)
	}

	paths := []string{}
	err = filepath.WalkDir(base, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("artifact path %s must not be a symlink", path)
		}
		if entry.IsDir() {
			return nil
		}
		if strings.ToLower(filepath.Ext(entry.Name())) != ".md" {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		paths = append(paths, filepath.ToSlash(relative))
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scan %s: %w", directory, err)
	}
	sort.Strings(paths)
	return paths, nil
}

func decodeFrontmatter(metadata map[string]any, target any, strict bool) error {
	data, err := yaml.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("encode frontmatter: %w", err)
	}
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(strict)
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("invalid frontmatter: %w", err)
	}
	return nil
}
