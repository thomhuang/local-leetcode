package util

import (
	"os"
	"strings"
)

type Log struct {
	Records []string
	Length  int64
}

func NewLog() *Log {
	return &Log{}
}

func (l *Log) OutputLogFile() {
	if l.Length > 0 {
		logFile, _ := os.Create("./output/local_leetcode_logs.txt")
		defer logFile.Close()

		_, _ = logFile.Write([]byte(strings.Join(l.Records, "\n")))
	}
}

func (l *Log) Append(message string) {
	l.Length++
	l.Records = append(l.Records, message)
}
