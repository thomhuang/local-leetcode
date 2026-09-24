package main

import (
	"strings"
	"testing"

	"github.com/thomhuang/local-leetcode/internal/solution"
)

func TestOutputSubmissionResultsAcceptedUsesStatusCode(t *testing.T) {
	runtimePercentile := 99.5
	memoryPercentile := 98.25
	response := solution.CheckSolutionResponse{
		StatusCode:        acceptedStatusCode,
		StatusRuntime:     "2 ms",
		StatusMemory:      "8.2 MB",
		TotalCorrect:      57,
		TotalTestcases:    57,
		RuntimePercentile: &runtimePercentile,
		MemoryPercentile:  &memoryPercentile,
	}

	got := OutputSubmissionResults(response)
	for _, want := range []string{
		"Accepted!",
		"Runtime: 2 ms",
		"Memory: 8.2 MB",
		"Test cases passed: 57/57",
		"Runtime percentile: 99.50%",
		"Memory percentile: 98.25%",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("OutputSubmissionResults() = %q, want it to contain %q", got, want)
		}
	}
}

func TestOutputSubmissionResultsWrongAnswerUsesSubmissionFields(t *testing.T) {
	response := solution.CheckSolutionResponse{
		StatusCode:     11,
		RunSuccessful:  true,
		StatusMessage:  "Wrong Answer",
		TotalCorrect:   18,
		TotalTestcases: 57,
		LastTestcase:   "[2,7,11,15]\n9",
		ExpectedOutput: "[0,1]",
	}

	got := OutputSubmissionResults(response)
	for _, want := range []string{
		"Wrong Answer",
		"Test cases passed: 18/57",
		"Input:\n[2,7,11,15]\n9",
		"Expected output:\n[0,1]",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("OutputSubmissionResults() = %q, want it to contain %q", got, want)
		}
	}
	if strings.Contains(got, "Answers:") {
		t.Errorf("OutputSubmissionResults() should not print an empty Answers section: %q", got)
	}
}

func TestOutputSubmissionResultsRunStyleAnswerIncludesMissingValues(t *testing.T) {
	response := solution.CheckSolutionResponse{
		StatusCode:         11,
		RunSuccessful:      true,
		CodeAnswer:         solution.StringList{"", "[1, 2]"},
		ExpectedCodeAnswer: solution.StringList{"[]", "[0, 1]"},
	}

	got := OutputSubmissionResults(response)
	for _, want := range []string{
		`Answers:`,
		`[1] Expected: "[]"`,
		`Actual:   ""`,
		`[2] Expected: "[0, 1]"`,
		`Actual:   "[1, 2]"`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("OutputSubmissionResults() = %q, want it to contain %q", got, want)
		}
	}
}
