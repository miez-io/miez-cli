package team

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/manuel/miez-cli/internal/lockfile"
	"github.com/manuel/miez-cli/internal/modules"
)

type distributionMove struct {
	from string
	to   string
}

type distributionMigration struct {
	moves     []distributionMove
	finalized bool
}

// prepareLegacyDistribution moves the previous root-level distribution state
// below .miez while keeping enough information to restore it if installation
// fails. Existing destination state is never overwritten.
func (s *Service) prepareLegacyDistribution() (*distributionMigration, error) {
	migration := &distributionMigration{moves: []distributionMove{}}
	moves := []distributionMove{
		{
			from: filepath.Join(s.Root, "miez_modules"),
			to:   modules.Root(s.Root),
		},
		{
			from: filepath.Join(s.Root, lockfile.FileName),
			to:   lockfile.Path(s.Root),
		},
	}
	for _, move := range moves {
		moved, err := moveLegacyPath(move)
		if err != nil {
			_ = migration.Rollback()
			return nil, err
		}
		if moved {
			migration.moves = append(migration.moves, move)
		}
	}
	return migration, nil
}

func moveLegacyPath(move distributionMove) (bool, error) {
	info, err := os.Lstat(move.from)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("inspect legacy distribution path %s: %w", move.from, err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return false, fmt.Errorf("legacy distribution path %s must not be a symlink", move.from)
	}
	if _, err := os.Lstat(move.to); err == nil {
		return false, fmt.Errorf("cannot migrate %s because %s already exists", move.from, move.to)
	} else if !errors.Is(err, os.ErrNotExist) {
		return false, fmt.Errorf("inspect distribution destination %s: %w", move.to, err)
	}
	if err := os.MkdirAll(filepath.Dir(move.to), 0o755); err != nil {
		return false, fmt.Errorf("create distribution directory: %w", err)
	}
	if err := os.Rename(move.from, move.to); err != nil {
		return false, fmt.Errorf("move legacy distribution path %s: %w", move.from, err)
	}
	return true, nil
}

func (migration *distributionMigration) Commit() {
	migration.finalized = true
}

func (migration *distributionMigration) Rollback() error {
	if migration == nil || migration.finalized {
		return nil
	}
	var rollbackErrors []error
	for index := len(migration.moves) - 1; index >= 0; index-- {
		move := migration.moves[index]
		if err := os.Rename(move.to, move.from); err != nil {
			rollbackErrors = append(rollbackErrors, fmt.Errorf("restore legacy distribution path %s: %w", move.from, err))
		}
	}
	migration.finalized = true
	return errors.Join(rollbackErrors...)
}

func removeEmptyLegacyState(root string) {
	paths := []string{
		filepath.Join(root, ".miez", "runs"),
		filepath.Join(root, ".miez", "changes", "archive"),
		filepath.Join(root, ".miez", "changes"),
	}
	for _, path := range paths {
		_ = os.Remove(path)
	}
}
