package util

import (
	"fmt"
	"os"
	"strings"

	q "github.com/thomhuang/local-leetcode/internal/question"
	s "github.com/thomhuang/local-leetcode/internal/solution"
)

func SaveMarkdownContent(ques q.Question) error {
	var sb strings.Builder

	sb.WriteString("# ")
	sb.WriteString(ques.Difficulty)
	sb.WriteString(": ")
	sb.WriteString(ques.QuestionId)
	sb.WriteString(". ")
	sb.WriteString(ques.Title)
	sb.WriteByte('\n')
	sb.WriteString(ques.Content)

	dir := "output/problems/" + ques.TitleSlug + "/"
	err := os.MkdirAll(dir, os.ModeDir)
	if err != nil {
		return err
	}

	err = os.WriteFile(dir+ques.TitleSlug+".md", []byte(sb.String()), os.ModePerm)
	if err != nil {
		return err
	}
	err = os.WriteFile(dir+ques.QuestionId+"-"+ques.TitleSlug+".go", []byte(ques.CodeSnippet), os.ModePerm)
	if err != nil {
		return err
	}

	fmt.Printf("Successfully created content for %s. %s\n", ques.QuestionId, ques.Title)
	return nil
}

func OutputQuestionResults(submission s.CheckSolutionResponse) string {
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
