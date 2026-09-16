package manifest_test

import (
	"testing"

	"github.com/manuel/miez-cli/internal/manifest"
)

func TestLoadMissingManifestReturnsEmpty(t *testing.T) {
	got, err := manifest.Load(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Teams) != 0 {
		t.Fatalf("Teams = %#v, want empty", got.Teams)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	root := t.TempDir()
	value := manifest.Manifest{Teams: map[string]string{"scrum": "https://github.com/manuel/miez-scrum-team"}}
	if err := manifest.Save(root, value); err != nil {
		t.Fatal(err)
	}
	got, err := manifest.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	url, ok := got.Resolve("scrum")
	if !ok || url != "https://github.com/manuel/miez-scrum-team" {
		t.Fatalf("Resolve(scrum) = (%q, %v)", url, ok)
	}
	if _, ok := got.Resolve("unknown"); ok {
		t.Fatal("Resolve accepted an unregistered name")
	}
}
