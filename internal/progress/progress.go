package progress

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

// Status is the persisted state of a problem. Only an accepted LeetCode
// submission is stored; the "written" state is inferred from the source file.
type Status string

const Accepted Status = "accepted"

type Record struct {
	Status Status    `json:"status"`
	At     time.Time `json:"at"`
}

// Store keeps accepted-problem records keyed by title slug.
type Store struct {
	path    string
	records map[string]Record
}

// Load reads the progress file. A missing or empty file yields an empty store.
func Load(path string) (*Store, error) {
	store := &Store{path: path, records: make(map[string]Record)}

	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return store, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read progress file: %w", err)
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return store, nil
	}
	if err := json.Unmarshal(data, &store.records); err != nil {
		return nil, fmt.Errorf("parse progress file: %w", err)
	}
	return store, nil
}

// Accepted reports whether a problem has an accepted record.
func (s *Store) Accepted(slug string) bool {
	record, ok := s.records[slug]
	return ok && record.Status == Accepted
}

// MarkAccepted records an accepted submission for a problem.
func (s *Store) MarkAccepted(slug string, at time.Time) {
	s.records[slug] = Record{Status: Accepted, At: at}
}

// Clear removes any record for a problem.
func (s *Store) Clear(slug string) {
	delete(s.records, slug)
}

// Len reports how many problems have a record.
func (s *Store) Len() int {
	return len(s.records)
}

// Save writes the store to disk.
func (s *Store) Save() error {
	data, err := json.MarshalIndent(s.records, "", "  ")
	if err != nil {
		return fmt.Errorf("encode progress: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return fmt.Errorf("create progress directory: %w", err)
	}
	if err := os.WriteFile(s.path, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write progress file: %w", err)
	}
	return nil
}
