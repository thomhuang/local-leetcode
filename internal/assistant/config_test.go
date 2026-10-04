package assistant

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadCreatesDefaults(t *testing.T) {
	dir := t.TempDir()

	cfg, prompt, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Model != DefaultModel {
		t.Fatalf("model = %q, want %q", cfg.Model, DefaultModel)
	}
	if cfg.ReasoningEffort != DefaultReasoningEffort {
		t.Fatalf("reasoning effort = %q, want %q", cfg.ReasoningEffort, DefaultReasoningEffort)
	}
	if !cfg.Thinking {
		t.Fatalf("thinking should default to enabled")
	}
	if !strings.Contains(prompt, "tutor") {
		t.Fatalf("default prompt unexpected: %q", prompt)
	}

	for _, name := range []string{configFileName, promptFileName} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Fatalf("expected %s to be created: %v", name, err)
		}
	}
}

func TestLoadCustomPromptAndConfig(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, promptFileName), []byte("Be terse."), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, configFileName), []byte(`{"model":"deepseek-v4-pro","thinking":false}`), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, prompt, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if prompt != "Be terse." {
		t.Fatalf("prompt = %q, want custom", prompt)
	}
	if cfg.Model != "deepseek-v4-pro" {
		t.Fatalf("model = %q", cfg.Model)
	}
	if cfg.Thinking {
		t.Fatalf("thinking should be false for the custom config")
	}
	// Fields absent from the file keep their defaults.
	if cfg.BaseURL != DefaultBaseURL {
		t.Fatalf("base url = %q, want default %q", cfg.BaseURL, DefaultBaseURL)
	}
}

func TestSaveConfigRoundTrip(t *testing.T) {
	dir := t.TempDir()
	want := DefaultConfig()
	want.Model = "deepseek-v4-pro"
	want.MaxTokens = 2048

	if err := SaveConfig(dir, want); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}
	got, _, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got != want {
		t.Fatalf("round trip = %+v, want %+v", got, want)
	}
}
