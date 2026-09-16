package model

import "regexp"

var identifierPattern = regexp.MustCompile(`^[a-z][a-z0-9]*(?:-[a-z0-9]+)*$`)

// IsValidIdentifier reports whether value is a kebab-case identifier, the
// shared format for team, worker, workflow, phase, and skill ids.
func IsValidIdentifier(value string) bool {
	return identifierPattern.MatchString(value)
}
