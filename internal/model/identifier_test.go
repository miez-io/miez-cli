package model

import "testing"

func TestIsValidIdentifier(t *testing.T) {
	valid := []string{"a", "team", "spec-driven-team", "a1", "a-1-b"}
	for _, value := range valid {
		if !IsValidIdentifier(value) {
			t.Errorf("IsValidIdentifier(%q) = false, want true", value)
		}
	}

	invalid := []string{"", "Team", "-team", "team-", "team--two", "team_two", "1team"}
	for _, value := range invalid {
		if IsValidIdentifier(value) {
			t.Errorf("IsValidIdentifier(%q) = true, want false", value)
		}
	}
}
