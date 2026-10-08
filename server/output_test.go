package main

import (
	"strings"
	"testing"

	"github.com/thomhuang/local-leetcode/internal/solution"
)

func TestOutputQuestionResultsUnsuccessfulPrintsRuntimeError(t *testing.T) {
	response := solution.CheckSolutionResponse{
		RunSuccessful:   false,
		StatusMessage:   "Runtime Error",
		FullRuntimeError: "panic: runtime error: index out of range [5] with length 3",
	}

	got := OutputQuestionResults(response)
	for _, want := range []string{
		"Unsuccessful run!",
		"Runtime Error",
		"Runtime error:\npanic: runtime error: index out of range [5] with length 3",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("OutputQuestionResults() = %q, want it to contain %q", got, want)
		}
	}
}

func TestOutputQuestionResultsUnsuccessfulFallsBackToRuntimeError(t *testing.T) {
	response := solution.CheckSolutionResponse{
		RunSuccessful: false,
		StatusMessage: "Runtime Error",
		RuntimeError:  "runtime error: index out of range",
	}

	got := OutputQuestionResults(response)
	if !strings.Contains(got, "Runtime error:\nruntime error: index out of range") {
		t.Errorf("OutputQuestionResults() = %q, want it to contain the runtime error", got)
	}
}

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
