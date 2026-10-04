package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/thomhuang/local-leetcode/internal/assistant"
)

// sseReader reads server-sent events that carry JSON chat chunks. DeepSeek
// streams one event per line ("data: {...}") and terminates with
// "data: [DONE]".
type sseReader struct {
	scanner *bufio.Scanner
}

func newSSEReader(r io.Reader) *sseReader {
	scanner := bufio.NewScanner(r)
	// A single delta is small, but leave room for tool-call payloads.
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	return &sseReader{scanner: scanner}
}

// Next returns the next JSON data event. done is true at the [DONE] sentinel or
// at the end of the stream.
func (s *sseReader) Next() (chunk assistant.StreamChunk, done bool, err error) {
	for s.scanner.Scan() {
		line := strings.TrimSpace(s.scanner.Text())
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}

		data, ok := strings.CutPrefix(line, "data:")
		if !ok {
			continue
		}
		data = strings.TrimSpace(data)
		if data == "" {
			continue
		}
		if data == "[DONE]" {
			return assistant.StreamChunk{}, true, nil
		}

		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			return assistant.StreamChunk{}, false, fmt.Errorf("parse stream chunk: %w", err)
		}
		return chunk, false, nil
	}

	if err := s.scanner.Err(); err != nil {
		return assistant.StreamChunk{}, false, err
	}
	return assistant.StreamChunk{}, true, nil
}
