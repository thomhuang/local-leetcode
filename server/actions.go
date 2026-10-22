package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/thomhuang/local-leetcode/internal/question"
	"github.com/thomhuang/local-leetcode/internal/user"
)

// loadUserCode reads a problem's solution file and prepares it for submission
// by stripping the local package clause.
func (app *App) loadUserCode(id int) (string, string, error) {
	meta, ok := app.Questions[id]
	if !ok {
		app.log().Error("code.load", "problem not in list", "problem", id)
		return "", "", fmt.Errorf("problem %d is not in the problem list", id)
	}

	titleSlug := meta.QuestionTitleSlug
	filePath := problemsDir + "/" + titleSlug + "/" + strconv.Itoa(id) + "-" + titleSlug + ".go"

	fileStream, err := os.ReadFile(filePath)
	if err != nil {
		app.log().Error("code.load", "read solution failed", "problem", id, "path", filePath, "error", err.Error())
		return "", "", fmt.Errorf("read %s: %w", filePath, err)
	}
	app.log().Debug("code.load", "read solution", "problem", id, "path", filePath, "bytes", len(fileStream))

	packageName := question.PackageName(titleSlug)
	return prepareSubmission(string(fileStream), packageName), titleSlug, nil
}

// TestProblem runs a problem's example test cases and returns the formatted
// result. It is safe to call outside the interactive prompt.
func (app *App) TestProblem(id int) (string, error) {
	app.log().Info("test.start", "running example tests", "problem", id)

	userSubmission, titleSlug, err := app.loadUserCode(id)
	if err != nil {
		return "", err
	}

	pendingSolution, err := app.fetchInterpretation(id, userSubmission)
	if err != nil {
		app.log().Error("test.fail", "submit for test run failed", "problem", id, "error", err.Error())
		return "", fmt.Errorf("submit code for a test run: %w", err)
	}

	result, err := app.pollSolution(pendingSolution.InterpretId, titleSlug)
	if err != nil {
		app.log().Error("test.fail", "poll test result failed", "problem", id, "error", err.Error())
		return "", fmt.Errorf("get the run result: %w", err)
	}
	app.log().Info("test.done", "example tests finished",
		"problem", id,
		"run_successful", result.RunSuccessful,
		"correct", result.CorrectAnswer)
	return OutputQuestionResults(result), nil
}

// SubmitProblem submits a problem and returns the formatted result. An
// accepted submission is recorded in the progress store.
func (app *App) SubmitProblem(id int) (string, error) {
	app.log().Info("submit.start", "submitting solution", "problem", id)

	userSubmission, titleSlug, err := app.loadUserCode(id)
	if err != nil {
		return "", err
	}

	pendingSolution, err := app.fetchSubmission(id, userSubmission)
	if err != nil {
		app.log().Error("submit.fail", "submit request failed", "problem", id, "error", err.Error())
		return "", fmt.Errorf("submit code for a submission: %w", err)
	}

	submissionId := strconv.FormatInt(pendingSolution.SubmissionId, 10)

	result, err := app.pollSolution(submissionId, titleSlug)
	if err != nil {
		app.log().Error("submit.fail", "poll submission result failed", "problem", id, "error", err.Error())
		return "", fmt.Errorf("get the submission result: %w", err)
	}

	if result.StatusCode == acceptedStatusCode || result.CorrectAnswer {
		if err := app.setAccepted(titleSlug, true); err != nil {
			app.log().Error("submit.progress", "unable to record accepted", "slug", titleSlug, "error", err.Error())
		}
	}

	app.log().Info("submit.done", "submission finished",
		"problem", id,
		"status_code", result.StatusCode,
		"correct", result.CorrectAnswer,
		"passed", result.TotalCorrect,
		"total", result.TotalTestcases)
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

// ResetProblem restores a problem's solution file to the starter stub fetched
// fresh from LeetCode, and clears its accepted record so the problem is fully
// re-opened.
func (app *App) ResetProblem(problem question.ProblemSetProblem) (string, error) {
	app.log().Info("problem.reset", "resetting solution", "id", problem.ID, "slug", problem.Slug)

	ques, err := app.fetchQuestionWithRetry(problem.Slug)
	if err != nil {
		app.log().Error("problem.reset", "fetch failed", "slug", problem.Slug, "error", err.Error())
		return "", fmt.Errorf("fetch question: %w", err)
	}
	if !hasProblemContent(ques) {
		app.log().Warn("problem.reset", "no starter code", "slug", problem.Slug)
		return "", fmt.Errorf("LeetCode returned no starter code for %s", problem.Slug)
	}

	target := starterFilePath(problemsDir, problem)
	if target == "" {
		target = problemsDir + "/" + ques.TitleSlug + "/" + ques.FrontEndQuestionId + "-" + ques.TitleSlug + ".go"
	}
	if err := os.WriteFile(target, []byte(question.MakeStarterCompilable(ques.CodeSnippet)), filePerm); err != nil {
		app.log().Error("problem.reset", "write starter failed", "slug", problem.Slug, "path", target, "error", err.Error())
		return "", fmt.Errorf("write starter file: %w", err)
	}

	if err := app.setAccepted(problem.Slug, false); err != nil {
		app.log().Error("problem.reset", "clear accepted failed", "slug", problem.Slug, "error", err.Error())
		return "", fmt.Errorf("clear accepted record: %w", err)
	}

	app.log().Info("problem.reset", "solution reset", "slug", problem.Slug, "path", target)
	return fmt.Sprintf("Reset %s. %s to starter", ques.FrontEndQuestionId, ques.Title), nil
}

// AddQuestion fetches a single problem and saves its statement and starter file.
func (app *App) AddQuestion(id int) (string, error) {
	meta, ok := app.Questions[id]
	if !ok {
		app.log().Warn("question.add", "problem not in list", "problem", id)
		return "", fmt.Errorf("problem %d was not found in the problem list", id)
	}

	app.log().Info("question.add", "fetching problem", "problem", id, "slug", meta.QuestionTitleSlug)
	ques, err := app.fetchQuestion(meta.QuestionTitleSlug)
	if err != nil {
		app.log().Error("question.add", "fetch failed", "problem", id, "error", err.Error())
		return "", fmt.Errorf("fetch question: %w", err)
	}
	if err := SaveMarkdownContent(ques); err != nil {
		app.log().Error("question.add", "save failed", "problem", id, "error", err.Error())
		return "", fmt.Errorf("save question content: %w", err)
	}
	app.log().Info("question.add", "problem added", "problem", id, "title", ques.Title)
	return fmt.Sprintf("Added %s. %s", ques.FrontEndQuestionId, ques.Title), nil
}

// Authenticate saves the cookies and verifies them against LeetCode.
func (app *App) Authenticate(session, csrfToken string) (user.UserStatusResponse, error) {
	app.log().Info("auth.save", "saving session cookies")
	if err := app.SaveAuthentication(session, csrfToken); err != nil {
		app.log().Error("auth.save", "save failed", "error", err.Error())
		return user.UserStatusResponse{}, fmt.Errorf("save authentication cookies: %w", err)
	}

	status, err := app.fetchUser()
	if err != nil {
		app.log().Error("auth.verify", "LeetCode rejected cookies", "error", err.Error())
		return user.UserStatusResponse{}, fmt.Errorf("LeetCode rejected the cookies: %w", err)
	}
	app.log().Info("auth.verify", "signed in", "username", status.Data.UserStatus.Username)
	return status, nil
}
