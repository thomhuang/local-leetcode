package util

import (
	"os"
	"time"
)

type Log struct {
	file *os.File
}

// NewLog opens the log file and truncates it. Each Append writes at once, so
// the log survives a panic or an interrupt. If the file cannot open, Append
// drops messages.
func NewLog(path string) *Log {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return &Log{}
	}
	return &Log{file: file}
}

func (l *Log) Append(message string) {
	if l.file == nil {
		return
	}
	_, _ = l.file.WriteString(time.Now().Format(time.RFC3339) + " " + message + "\n")
}

func (l *Log) Close() {
	if l.file != nil {
		_ = l.file.Close()
	}
}
