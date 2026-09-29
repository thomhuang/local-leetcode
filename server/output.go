package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/thomhuang/local-leetcode/internal/question"
	"github.com/thomhuang/local-leetcode/internal/solution"
)

func SaveMarkdownContent(ques question.Question) error {
	if !hasProblemContent(ques) {
		return SaveProblemPlaceholder(ques)
	}

	var sb strings.Builder

	sb.WriteString("# ")
	sb.WriteString(ques.Difficulty)
	sb.WriteString(": ")
	sb.WriteString(ques.FrontEndQuestionId)
	sb.WriteString(". ")
	sb.WriteString(ques.Title)
	sb.WriteByte('\n')
	sb.WriteString(ques.Content)

	dir := problemsDir + "/" + ques.TitleSlug + "/"
	err := os.MkdirAll(dir, dirPerm)
	if err != nil {
		return err
	}

	err = os.WriteFile(dir+ques.TitleSlug+".md", []byte(sb.String()), filePerm)
	if err != nil {
		return err
	}
	err = os.WriteFile(dir+ques.FrontEndQuestionId+"-"+ques.TitleSlug+".go", []byte(question.MakeStarterCompilable(ques.CodeSnippet)), filePerm)
	if err != nil {
		return err
	}

	fmt.Printf("Successfully created content for %s. %s\n", ques.FrontEndQuestionId, ques.Title)
	return nil
}

// hasProblemContent reports whether LeetCode returned a usable statement and
// Go starter code. Premium problems come back with both fields empty.
func hasProblemContent(ques question.Question) bool {
	return strings.TrimSpace(ques.Content) != "" && strings.TrimSpace(ques.CodeSnippet) != ""
}

// SaveProblemPlaceholder records a problem whose content LeetCode would not
// return. Premium problems have no statement or starter code for
// non-subscribers, so a placeholder test file is written instead.
func SaveProblemPlaceholder(ques question.Question) error {
	dir := problemsDir + "/" + ques.TitleSlug + "/"
	if err := os.MkdirAll(dir, dirPerm); err != nil {
		return err
	}

	reason := "LeetCode did not return a problem statement or starter code"
	if ques.IsPaidOnly {
		reason = "This is a LeetCode Premium problem; its statement is not available to non-subscribers"
	}

	var sb strings.Builder
	sb.WriteString("# ")
	sb.WriteString(ques.Difficulty)
	sb.WriteString(": ")
	sb.WriteString(ques.FrontEndQuestionId)
	sb.WriteString(". ")
	sb.WriteString(ques.Title)
	sb.WriteString("\n\n> ")
	sb.WriteString(reason)
	sb.WriteString(". A placeholder test file was created instead of a starter file.\n")

	err := os.WriteFile(dir+ques.TitleSlug+".md", []byte(sb.String()), filePerm)
	if err != nil {
		return err
	}

	packageName := question.PackageName(ques.TitleSlug)
	var test strings.Builder
	fmt.Fprintf(&test, "package %s\n\n", packageName)
	fmt.Fprintf(&test, "import \"testing\"\n\n")
	fmt.Fprintf(&test, "// %s. %s. Replace this placeholder once the statement and starter code are available.\n",
		ques.FrontEndQuestionId, reason)
	fmt.Fprintf(&test, "func TestNotImported(t *testing.T) {\n\tt.Skip(%q)\n}\n", reason)

	err = os.WriteFile(dir+ques.TitleSlug+"_test.go", []byte(test.String()), filePerm)
	if err != nil {
		return err
	}

	fmt.Printf("Created placeholder for %s. %s (%s)\n", ques.FrontEndQuestionId, ques.Title, reason)
	return nil
}

func OutputQuestionResults(submission solution.CheckSolutionResponse) string {
	var sb strings.Builder

	sb.WriteRune('\n')
	runSuccessful := submission.RunSuccessful
	if runSuccessful {
		sb.WriteString("Accepted, ")
		if submission.CorrectAnswer {
			sb.WriteString("Correct!\n\n")
			sb.WriteString("Runtime: ")
			sb.WriteString(submission.StatusRuntime)
			sb.WriteString("\n\n")
		} else {
			sb.WriteString("Incorrect :(\n\n")

			n := len(submission.ExpectedCodeAnswer)
			sb.WriteString("Answers: \n")
			for i := range n {
				if len(submission.ExpectedCodeAnswer[i]) == 0 {
					continue
				}

				sb.WriteString("Expected: ")
				sb.WriteString(submission.ExpectedCodeAnswer[i])
				sb.WriteString("\t\t")
				sb.WriteString("Actual: ")
				sb.WriteString(submission.CodeAnswer[i])
				sb.WriteRune('\n')
			}

			m := len(submission.ExpectedCodeOutput)
			sb.WriteString("Code Output: \n")
			for j := range m {
				if len(submission.ExpectedCodeOutput[j]) == 0 {
					continue
				}

				sb.WriteString("Expected: ")
				sb.WriteString(submission.ExpectedCodeOutput[j])
				sb.WriteString("\t\t")
				sb.WriteString("Actual: ")
				sb.WriteString(submission.CodeOutput[j])
				sb.WriteRune('\n')
			}

			z := len(submission.ExpectedStdOutputList)
			sb.WriteString("Std Output: \n")
			for k := range z {
				if len(submission.ExpectedStdOutputList[k]) == 0 {
					continue
				}

				sb.WriteString("Expected: ")
				sb.WriteString(submission.ExpectedStdOutputList[k])
				sb.WriteString("\t\t")
				sb.WriteString("Actual: ")
				sb.WriteString(submission.StdOutputList[k])
				sb.WriteRune('\n')
			}
		}
	} else {
		sb.WriteString("Unsuccessful run!\n")
		sb.WriteString(submission.StatusMessage)
		sb.WriteRune('\n')
		if len(submission.CompileError) != 0 {
			sb.WriteString(submission.CompileError)
		}
	}

	return sb.String()
}

const acceptedStatusCode = 10

func OutputSubmissionResults(submission solution.CheckSolutionResponse) string {
	var sb strings.Builder
	sb.WriteByte('\n')

	// Full submissions do not consistently return correct_answer. The judge
	// status code is authoritative and is 10 for an accepted submission.
	if submission.StatusCode == acceptedStatusCode || submission.CorrectAnswer {
		sb.WriteString("Accepted!\n\n")
		if submission.StatusRuntime != "" {
			sb.WriteString("Runtime: ")
			sb.WriteString(submission.StatusRuntime)
			sb.WriteByte('\n')
		}
		if submission.StatusMemory != "" {
			sb.WriteString("Memory: ")
			sb.WriteString(submission.StatusMemory)
			sb.WriteByte('\n')
		}
		sb.WriteString(fmt.Sprintf("Test cases passed: %d/%d\n", submission.TotalCorrect, submission.TotalTestcases))
		if submission.RuntimePercentile != nil {
			sb.WriteString(fmt.Sprintf("Runtime percentile: %.2f%%\n", *submission.RuntimePercentile))
		}
		if submission.MemoryPercentile != nil {
			sb.WriteString(fmt.Sprintf("Memory percentile: %.2f%%\n", *submission.MemoryPercentile))
		}
		return sb.String()
	}

	statusMessage := strings.TrimSpace(submission.StatusMessage)
	if statusMessage == "" {
		statusMessage = "Submission failed!"
	}
	sb.WriteString(statusMessage)
	sb.WriteString("\n\n")

	if submission.RunSuccessful {
		sb.WriteString(fmt.Sprintf("Test cases passed: %d/%d\n\n", submission.TotalCorrect, submission.TotalTestcases))
	}

	if submission.CompileError != "" {
		sb.WriteString("Compile error:\n")
		sb.WriteString(submission.CompileError)
		sb.WriteByte('\n')
	}
	runtimeError := submission.RuntimeError
	if runtimeError == "" {
		runtimeError = submission.FullRuntimeError
	}
	if runtimeError != "" {
		sb.WriteString("Runtime error:\n")
		sb.WriteString(runtimeError)
		sb.WriteByte('\n')
	}

	// A full submission reports the failing input and expected output in these
	// fields. It does not include the code_answer fields used by test runs.
	failingInput := submission.LastTestcase
	if failingInput == "" {
		failingInput = submission.InputFormatted
	}
	if failingInput == "" {
		failingInput = submission.Input
	}
	if failingInput != "" {
		sb.WriteString("Input:\n")
		sb.WriteString(failingInput)
		if !strings.HasSuffix(failingInput, "\n") {
			sb.WriteByte('\n')
		}
	}
	if submission.ExpectedOutput != "" {
		sb.WriteString("Expected output:\n")
		sb.WriteString(submission.ExpectedOutput)
		if !strings.HasSuffix(submission.ExpectedOutput, "\n") {
			sb.WriteByte('\n')
		}
	}
	if submission.StdOutput != "" {
		sb.WriteString("Standard output:\n")
		sb.WriteString(submission.StdOutput)
		if !strings.HasSuffix(submission.StdOutput, "\n") {
			sb.WriteByte('\n')
		}
	}

	// Keep support for run-style payloads in case the endpoint returns them.
	writeAnswerComparisons(&sb, "Answers", submission.ExpectedCodeAnswer, submission.CodeAnswer)
	writeAnswerComparisons(&sb, "Code output", submission.ExpectedCodeOutput, submission.CodeOutput)
	writeAnswerComparisons(&sb, "Standard output", submission.ExpectedStdOutputList, submission.StdOutputList)

	return sb.String()
}

func writeAnswerComparisons(sb *strings.Builder, title string, expected, actual solution.StringList) {
	count := len(expected)
	if len(actual) > count {
		count = len(actual)
	}
	if count == 0 {
		return
	}

	sb.WriteString(title)
	sb.WriteString(":\n")
	for i := range count {
		expectedValue := "<missing>"
		if i < len(expected) {
			expectedValue = expected[i]
		}
		actualValue := "<missing>"
		if i < len(actual) {
			actualValue = actual[i]
		}
		sb.WriteString(fmt.Sprintf("  [%d] Expected: %q\n", i+1, expectedValue))
		sb.WriteString(fmt.Sprintf("       Actual:   %q\n", actualValue))
	}
}
