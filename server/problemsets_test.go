package main

import (
	"testing"

	"github.com/thomhuang/local-leetcode/internal/question"
)

func TestHasProblemContent(t *testing.T) {
	tests := []struct {
		name string
		ques question.Question
		want bool
	}{
		{
			name: "statement and starter code",
			ques: question.Question{Content: "# Two Sum", CodeSnippet: "package two_sum"},
			want: true,
		},
		{
			name: "premium problem is empty",
			ques: question.Question{IsPaidOnly: true},
			want: false,
		},
		{
			name: "whitespace does not count",
			ques: question.Question{Content: "  \n", CodeSnippet: "\t"},
			want: false,
		},
		{
			name: "missing starter code",
			ques: question.Question{Content: "# Two Sum"},
			want: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := hasProblemContent(test.ques); got != test.want {
				t.Fatalf("hasProblemContent() = %v, want %v", got, test.want)
			}
		})
	}
}
