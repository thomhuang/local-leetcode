package assistant

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadCreatesConfigDefault(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, promptFileName), []byte("tutor prompt"), 0o644); err != nil {
		t.Fatal(err)
	}

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
	if prompt != "tutor prompt" {
		t.Fatalf("prompt = %q, want custom", prompt)
	}
	if _, err := os.Stat(filepath.Join(dir, configFileName)); err != nil {
		t.Fatalf("expected %s to be created: %v", configFileName, err)
	}
}

func TestLoadMissingPromptErrors(t *testing.T) {
	dir := t.TempDir()
	if _, _, err := Load(dir); err == nil {
		t.Fatal("Load should fail when system_prompt.md is missing")
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
	if err := os.WriteFile(filepath.Join(dir, promptFileName), []byte("prompt"), 0o644); err != nil {
		t.Fatalf("write prompt: %v", err)
	}
	got, _, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got != want {
		t.Fatalf("round trip = %+v, want %+v", got, want)
	}
}
