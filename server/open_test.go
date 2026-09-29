package main

import (
	"strings"
	"testing"
)

func TestWindowsOpenCommand(t *testing.T) {
	path := `C:\Users\me\file.go`
	name, args := windowsOpenCommand(path)

	if name != "cmd" {
		t.Fatalf("windowsOpenCommand() name = %q, want cmd", name)
	}
	want := []string{"/c", "start", "", path}
	if len(args) != len(want) {
		t.Fatalf("windowsOpenCommand() args = %v, want %v", args, want)
	}
	for i := range want {
		if args[i] != want[i] {
			t.Fatalf("windowsOpenCommand() args = %v, want %v", args, want)
		}
	}
}

func TestOpenInDefaultAppEmptyPath(t *testing.T) {
	if err := openInDefaultApp(""); err == nil {
		t.Fatalf("openInDefaultApp(\"\") error = nil, want an error")
	}
}

func TestOpenFileWithoutImport(t *testing.T) {
	m := testBrowseModel()
	m.app = &App{}
	m.mode = modeDetail
	m.detailIndex = 0

	// The test working directory has no matching problem file, so opening it
	// should report that there is nothing to open rather than panic.
	m.openFile()
	if !strings.Contains(m.message, "No .go file") {
		t.Fatalf("openFile() message = %q, want a missing-file hint", m.message)
	}
}
