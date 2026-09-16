package team

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/manuel/miez-cli/internal/lockfile"
	"github.com/manuel/miez-cli/internal/modules"
)

// AuditFinding is one drift finding for a single deployed file.
type AuditFinding struct {
	Path   string
	Status string // "modified", "missing", or "extra"
}

const (
	StatusModified = "modified"
	StatusMissing  = "missing"
	StatusExtra    = "extra"
)

// AuditReport is one installed team's drift comparison against its lockfile
// entry.
type AuditReport struct {
	TeamID   string
	Findings []AuditFinding
}

// Clean reports whether report found no drift.
func (report AuditReport) Clean() bool {
	return len(report.Findings) == 0
}

// Audit compares teamID's (or, when empty, every installed team's) deployed
// files against the lockfile's recorded hashes, without any network access.
func (s *Service) Audit(teamID string) ([]AuditReport, error) {
	lock, err := lockfile.Load(s.Root)
	if err != nil {
		return nil, err
	}
	teamIDs := []string{}
	if teamID != "" {
		if _, ok := lock.Teams[teamID]; !ok {
			return nil, fmt.Errorf("team %q is not installed", teamID)
		}
		teamIDs = append(teamIDs, teamID)
	} else {
		for id := range lock.Teams {
			teamIDs = append(teamIDs, id)
		}
		sort.Strings(teamIDs)
	}

	reports := make([]AuditReport, 0, len(teamIDs))
	for _, id := range teamIDs {
		entry := lock.Teams[id]
		teamRoot, err := modules.TeamPath(s.Root, id)
		if err != nil {
			return nil, err
		}
		findings, err := auditTeam(teamRoot, entry)
		if err != nil {
			return nil, err
		}
		reports = append(reports, AuditReport{TeamID: id, Findings: findings})
	}
	return reports, nil
}

func auditTeam(teamRoot string, entry lockfile.Entry) ([]AuditFinding, error) {
	findings := []AuditFinding{}

	for _, relative := range entry.DeployedFiles {
		path := filepath.Join(teamRoot, filepath.FromSlash(relative))
		info, err := os.Stat(path)
		if os.IsNotExist(err) {
			findings = append(findings, AuditFinding{Path: relative, Status: StatusMissing})
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("inspect %s: %w", relative, err)
		}
		if info.IsDir() {
			findings = append(findings, AuditFinding{Path: relative, Status: StatusMissing})
			continue
		}
		hash, err := lockfile.HashFile(path)
		if err != nil {
			return nil, err
		}
		if hash != entry.FileHashes[relative] {
			findings = append(findings, AuditFinding{Path: relative, Status: StatusModified})
		}
	}

	declared := map[string]struct{}{}
	for _, relative := range entry.DeployedFiles {
		declared[relative] = struct{}{}
	}
	err := filepath.WalkDir(teamRoot, func(path string, entryInfo os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entryInfo.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(teamRoot, path)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		if _, ok := declared[relative]; !ok {
			findings = append(findings, AuditFinding{Path: relative, Status: StatusExtra})
		}
		return nil
	})
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	sort.Slice(findings, func(i, j int) bool {
		if findings[i].Path == findings[j].Path {
			return findings[i].Status < findings[j].Status
		}
		return findings[i].Path < findings[j].Path
	})
	return findings, nil
}
