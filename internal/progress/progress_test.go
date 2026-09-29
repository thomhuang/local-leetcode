package progress

import (
	"path/filepath"
	"testing"
	"time"
)

func TestLoadMissingFileIsEmpty(t *testing.T) {
	store, err := Load(filepath.Join(t.TempDir(), "progress.json"))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if store.Len() != 0 {
		t.Fatalf("Load() len = %d, want 0", store.Len())
	}
	if store.Accepted("two-sum") {
		t.Fatalf("Accepted() = true on an empty store")
	}
}

func TestMarkSaveReload(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "progress.json")
	store, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	at := time.Date(2026, 9, 28, 12, 0, 0, 0, time.UTC)
	store.MarkAccepted("two-sum", at)
	if !store.Accepted("two-sum") {
		t.Fatalf("Accepted() = false after MarkAccepted")
	}
	if err := store.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	reloaded, err := Load(path)
	if err != nil {
		t.Fatalf("reload error = %v", err)
	}
	if !reloaded.Accepted("two-sum") {
		t.Fatalf("reloaded store lost the record")
	}
	if reloaded.Accepted("valid-anagram") {
		t.Fatalf("reloaded store has an unexpected record")
	}
}

func TestClear(t *testing.T) {
	path := filepath.Join(t.TempDir(), "progress.json")
	store, _ := Load(path)
	store.MarkAccepted("two-sum", time.Now())
	store.Clear("two-sum")
	if store.Accepted("two-sum") {
		t.Fatalf("Accepted() = true after Clear")
	}
	if store.Len() != 0 {
		t.Fatalf("Len() = %d after Clear, want 0", store.Len())
	}
}
