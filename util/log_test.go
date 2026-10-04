package util

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func newTestLog(t *testing.T) *Log {
	t.Helper()
	log := NewLog(filepath.Join(t.TempDir(), "test.log"))
	t.Cleanup(log.Close)
	return log
}

func TestLogRecordsLevelsAndFields(t *testing.T) {
	log := newTestLog(t)

	log.Debug("dbg.event", "debug message", "count", 1)
	log.Info("inf.event", "info message")
	log.Warn("wrn.event", "warn message", "leftover")
	log.Error("err.event", "error message", "reason", "boom")

	entries := log.Entries()
	if len(entries) != 4 {
		t.Fatalf("entries = %d, want 4", len(entries))
	}

	wantLevels := []Level{LevelDebug, LevelInfo, LevelWarn, LevelError}
	for i, want := range wantLevels {
		if entries[i].Level != want {
			t.Fatalf("entry %d level = %v, want %v", i, entries[i].Level, want)
		}
	}

	if entries[0].Event != "dbg.event" || entries[0].Message != "debug message" {
		t.Fatalf("first entry = %+v", entries[0])
	}
	if len(entries[0].Fields) != 1 || entries[0].Fields[0].Key != "count" || entries[0].Fields[0].Value != "1" {
		t.Fatalf("first entry fields = %+v", entries[0].Fields)
	}
	// A trailing key without a value is kept.
	if len(entries[2].Fields) != 1 || entries[2].Fields[0].Value != "" {
		t.Fatalf("warn entry fields = %+v", entries[2].Fields)
	}
	if len(entries[3].Fields) != 1 || entries[3].Fields[0].Value != "boom" {
		t.Fatalf("error entry fields = %+v", entries[3].Fields)
	}
}

func TestLogAppendIsInfo(t *testing.T) {
	log := newTestLog(t)
	log.Append("plain")

	entries := log.Entries()
	if len(entries) != 1 || entries[0].Level != LevelInfo || entries[0].Message != "plain" {
		t.Fatalf("entries = %+v", entries)
	}
}

func TestLogVerboseGate(t *testing.T) {
	log := newTestLog(t)

	log.VerboseDebug("verbose.event", "hidden")
	if got := len(log.Entries()); got != 0 {
		t.Fatalf("verbose off recorded %d entries, want 0", got)
	}

	log.SetVerbose(true)
	if !log.Verbose() {
		t.Fatalf("Verbose() = false after SetVerbose(true)")
	}
	log.VerboseDebug("verbose.event", "shown")
	if got := len(log.Entries()); got != 1 {
		t.Fatalf("verbose on recorded %d entries, want 1", got)
	}
}

func TestLogRingBufferTrims(t *testing.T) {
	log := newTestLog(t)
	log.max = 3

	for i := 0; i < 5; i++ {
		log.Info("ring.event", "message")
	}

	entries := log.Entries()
	if len(entries) != 3 {
		t.Fatalf("entries = %d, want 3", len(entries))
	}
}

func TestLogClear(t *testing.T) {
	log := newTestLog(t)
	log.Info("event", "message")
	log.Clear()
	if got := len(log.Entries()); got != 0 {
		t.Fatalf("entries after Clear = %d, want 0", got)
	}
}

func TestLogWritesFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.log")
	log := NewLog(path)
	log.Info("http.request", "sending", "method", "GET")
	log.Close()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	line := string(data)
	for _, want := range []string{"INFO", "http.request", "sending", "method=GET"} {
		if !strings.Contains(line, want) {
			t.Fatalf("log file missing %q:\n%s", want, line)
		}
	}
}

func TestLogNilSafe(t *testing.T) {
	var log *Log

	log.Append("x")
	log.Debug("e", "m")
	log.Info("e", "m")
	log.Warn("e", "m")
	log.Error("e", "m")
	log.VerboseDebug("e", "m")
	log.SetVerbose(true)
	log.Clear()
	log.Close()

	if got := log.Entries(); got != nil {
		t.Fatalf("nil log Entries() = %+v, want nil", got)
	}
	if log.Verbose() {
		t.Fatalf("nil log Verbose() = true, want false")
	}
	if log.Path() != "" {
		t.Fatalf("nil log Path() = %q, want empty", log.Path())
	}
}
