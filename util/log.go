package util

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// Level is the severity of a log entry, ordered from most to least verbose.
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "INFO"
	}
}

// Field is one structured key/value pair attached to an entry.
type Field struct {
	Key   string
	Value string
}

// Entry is one recorded log line. It is copied out of the buffer, so a caller
// may keep it after the record call returns.
type Entry struct {
	Time    time.Time
	Level   Level
	Event   string
	Message string
	Fields  []Field
}

// Log records structured entries to an in-memory ring buffer and, when the file
// opens, appends them to disk. Every method is safe for concurrent use: the
// bubbletea command goroutines log while the UI reads snapshots. The buffer
// still works when the file cannot be opened, so the debug screen keeps working
// even if the log path is unwritable.
type Log struct {
	mu      sync.Mutex
	file    *os.File
	path    string
	entries []Entry
	max     int
	verbose bool
}

// defaultMaxEntries bounds the in-memory buffer so a long session cannot grow
// without limit.
const defaultMaxEntries = 2000

// NewLog opens the log file and truncates it. When the file cannot be opened the
// in-memory buffer is still available.
func NewLog(path string) *Log {
	log := &Log{path: path, max: defaultMaxEntries}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err == nil {
		log.file = file
	}
	return log
}

// Append records a plain informational message. It is kept for callers that do
// not care about an event name or structured fields.
func (l *Log) Append(message string) {
	l.log(LevelInfo, "log", message)
}

// Debug records a detailed entry, usually breadcrumbs that are only useful when
// diagnosing a problem.
func (l *Log) Debug(event, message string, kv ...any) {
	l.log(LevelDebug, event, message, kv...)
}

// Info records a normal lifecycle event.
func (l *Log) Info(event, message string, kv ...any) {
	l.log(LevelInfo, event, message, kv...)
}

// Warn records something unexpected that did not stop the operation.
func (l *Log) Warn(event, message string, kv ...any) {
	l.log(LevelWarn, event, message, kv...)
}

// Error records a failed operation.
func (l *Log) Error(event, message string, kv ...any) {
	l.log(LevelError, event, message, kv...)
}

// VerboseDebug records a potentially large entry only when verbose mode is on.
// Use it for full request/response bodies.
func (l *Log) VerboseDebug(event, message string, kv ...any) {
	if l == nil || !l.Verbose() {
		return
	}
	l.log(LevelDebug, event, message, kv...)
}

func (l *Log) log(level Level, event, message string, kv ...any) {
	if l == nil {
		return
	}

	entry := Entry{
		Time:    time.Now(),
		Level:   level,
		Event:   event,
		Message: message,
		Fields:  fieldsFrom(kv),
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	l.entries = append(l.entries, entry)
	if drop := len(l.entries) - l.max; drop > 0 {
		l.entries = append([]Entry(nil), l.entries[drop:]...)
	}

	if l.file != nil {
		_, _ = l.file.WriteString(entry.format() + "\n")
	}
}

// Entries returns a copy of the buffered entries so callers can render them
// without holding the lock.
func (l *Log) Entries() []Entry {
	if l == nil {
		return nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]Entry, len(l.entries))
	copy(out, l.entries)
	return out
}

// Clear empties the in-memory buffer. The log file is left untouched.
func (l *Log) Clear() {
	if l == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = nil
}

// SetVerbose turns recording of large payloads on or off.
func (l *Log) SetVerbose(verbose bool) {
	if l == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.verbose = verbose
}

// Verbose reports whether large payloads are being recorded.
func (l *Log) Verbose() bool {
	if l == nil {
		return false
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.verbose
}

// Path is the file the entries are mirrored to, which may be empty when the log
// has no backing file.
func (l *Log) Path() string {
	if l == nil {
		return ""
	}
	return l.path
}

// Close flushes and closes the backing file. The buffer remains readable.
func (l *Log) Close() {
	if l == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file != nil {
		_ = l.file.Close()
		l.file = nil
	}
}

func (e Entry) format() string {
	var b strings.Builder
	b.WriteString(e.Time.Format(time.RFC3339))
	b.WriteByte(' ')
	b.WriteString(e.Level.String())
	if e.Event != "" {
		b.WriteByte(' ')
		b.WriteString(e.Event)
	}
	if e.Message != "" {
		b.WriteByte(' ')
		b.WriteString(e.Message)
	}
	for _, field := range e.Fields {
		b.WriteByte(' ')
		b.WriteString(field.Key)
		b.WriteByte('=')
		b.WriteString(field.Value)
	}
	return b.String()
}

// fieldsFrom pairs the variadic key/value arguments into fields. A trailing key
// without a value is kept with an empty value.
func fieldsFrom(kv []any) []Field {
	if len(kv) == 0 {
		return nil
	}
	fields := make([]Field, 0, (len(kv)+1)/2)
	for i := 0; i < len(kv); i += 2 {
		field := Field{Key: fmt.Sprint(kv[i])}
		if i+1 < len(kv) {
			field.Value = fmt.Sprint(kv[i+1])
		}
		fields = append(fields, field)
	}
	return fields
}
