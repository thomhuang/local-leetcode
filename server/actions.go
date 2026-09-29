package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/thomhuang/local-leetcode/internal/question"
)

// loadUserCode reads a problem's solution file and prepares it for submission
// by stripping the local package clause.
func (app *App) loadUserCode(id int) (string, string, error) {
	meta, ok := app.Questions[id]
	if !ok {
		return "", "", fmt.Errorf("problem %d is not in the problem list", id)
	}

	titleSlug := meta.QuestionTitleSlug
	filePath := problemsDir + "/" + titleSlug + "/" + strconv.Itoa(id) + "-" + titleSlug + ".go"

	fileStream, err := os.ReadFile(filePath)
	if err != nil {
		return "", "", fmt.Errorf("read %s: %w", filePath, err)
	}

	packageName := question.PackageName(titleSlug)
	return prepareSubmission(string(fileStream), packageName), titleSlug, nil
}

// TestProblem runs a problem's example test cases and returns the formatted
// result. It is safe to call outside the interactive prompt.
func (app *App) TestProblem(id int) (string, error) {
	userSubmission, titleSlug, err := app.loadUserCode(id)
	if err != nil {
		return "", err
	}

	pendingSolution, err := app.fetchInterpretation(id, userSubmission)
	if err != nil {
		return "", fmt.Errorf("submit code for a test run: %w", err)
	}

	result, err := app.pollSolution(pendingSolution.InterpretId, titleSlug)
	if err != nil {
		return "", fmt.Errorf("get the run result: %w", err)
	}
	return OutputQuestionResults(result), nil
}

// SubmitProblem submits a problem and returns the formatted result. An
// accepted submission is recorded in the progress store.
func (app *App) SubmitProblem(id int) (string, error) {
	userSubmission, titleSlug, err := app.loadUserCode(id)
	if err != nil {
		return "", err
	}

	pendingSolution, err := app.fetchSubmission(id, userSubmission)
	if err != nil {
		return "", fmt.Errorf("submit code for a submission: %w", err)
	}

	submissionId := strconv.FormatInt(pendingSolution.SubmissionId, 10)

	result, err := app.pollSolution(submissionId, titleSlug)
	if err != nil {
		return "", fmt.Errorf("get the submission result: %w", err)
	}

	if result.StatusCode == acceptedStatusCode || result.CorrectAnswer {
		if err := app.setAccepted(titleSlug, true); err != nil {
			app.Log.Append(fmt.Sprintf("Unable to record progress for %s: %s", titleSlug, err.Error()))
		}
	}

	return OutputSubmissionResults(result), nil
}

// setAccepted stores or clears an accepted record for a problem.
func (app *App) setAccepted(slug string, accepted bool) error {
	if app.Progress == nil {
		return nil
	}
	if accepted {
		app.Progress.MarkAccepted(slug, time.Now())
	} else {
		app.Progress.Clear(slug)
	}
	return app.Progress.Save()
}
