package main

import (
	"fmt"
	"os/exec"
)

// windowsOpenCommand builds the command that opens a file with its associated
// default application on Windows. start treats its first argument as a window
// title, so an empty one is passed before the path.
func windowsOpenCommand(path string) (string, []string) {
	return "cmd", []string{"/c", "start", "", path}
}

// openInDefaultApp launches the default Windows application for a file and
// returns without waiting for it to exit.
func openInDefaultApp(path string) error {
	if path == "" {
		return fmt.Errorf("no file to open")
	}

	name, args := windowsOpenCommand(path)
	if err := exec.Command(name, args...).Start(); err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}
	return nil
}
