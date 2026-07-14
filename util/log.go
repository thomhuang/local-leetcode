package util

import (
	"os"
	"strings"
)

type Log struct {
	Records []string
}

func NewLog() *Log {
	return &Log{}
}

func (l *Log) OutputLogFile() {
	if len(l.Records) > 0 {
		logFile, _ := os.Create("server/output/local_leetcode_logs.txt")
		defer logFile.Close()

		_, _ = logFile.Write([]byte(strings.Join(l.Records, "\n")))
	}
}

func (l *Log) Append(message string) {
	l.Records = append(l.Records, message)
}
