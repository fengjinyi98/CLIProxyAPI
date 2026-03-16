package synthesizer

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/router-for-me/CLIProxyAPI/v6/internal/config"
)

func TestFileSynthesizer_AssignsAuthGroupFromConfiguredDirectory(t *testing.T) {
	root := t.TempDir()
	normalDir := filepath.Join(root, "auths-normal")
	teamDir := filepath.Join(root, "auths-team")
	if err := os.MkdirAll(normalDir, 0o755); err != nil {
		t.Fatalf("mkdir normal dir: %v", err)
	}
	if err := os.MkdirAll(teamDir, 0o755); err != nil {
		t.Fatalf("mkdir team dir: %v", err)
	}

	authPath := filepath.Join(teamDir, "team.json")
	if err := os.WriteFile(authPath, []byte(`{"type":"codex","email":"team@example.com"}`), 0o600); err != nil {
		t.Fatalf("write auth file: %v", err)
	}

	synth := NewFileSynthesizer()
	auths, err := synth.Synthesize(&SynthesisContext{
		Config: &config.Config{
			AuthDir: normalDir,
			AuthGroups: []config.ClientAuthGroup{
				{Name: "team", AuthDirs: []string{teamDir}},
			},
		},
		AuthDir:     teamDir,
		Now:         time.Now(),
		IDGenerator: NewStableIDGenerator(),
	})
	if err != nil {
		t.Fatalf("synthesize: %v", err)
	}
	if len(auths) != 1 {
		t.Fatalf("auths len = %d, want 1", len(auths))
	}
	if got := auths[0].AuthGroup(); got != "team" {
		t.Fatalf("auth group = %q, want %q", got, "team")
	}
}
