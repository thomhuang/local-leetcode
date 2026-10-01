package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
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

// addQuestionCmd fetches a single problem and saves its statement and starter
// file.
func (app *App) addQuestionCmd(id int) tea.Cmd {
	return func() tea.Msg {
		meta, ok := app.Questions[id]
		if !ok {
			return tuiResultMsg{title: "Add a new question", body: fmt.Sprintf("Problem %d was not found in the problem list.", id)}
		}

		ques, err := app.fetchQuestion(meta.QuestionTitleSlug)
		if err != nil {
			return tuiResultMsg{title: "Add a new question", body: "Failed to fetch question: " + err.Error()}
		}
		if err := SaveMarkdownContent(ques); err != nil {
			return tuiResultMsg{title: "Add a new question", body: "Failed to save question content: " + err.Error()}
		}
		return tuiResultMsg{
			title: "Add a new question",
			body:  fmt.Sprintf("Added %s. %s", ques.FrontEndQuestionId, ques.Title),
		}
	}
}

// importSetCmd imports every missing problem in a named set.
func (app *App) importSetCmd(name string) tea.Cmd {
	return func() tea.Msg {
		set, err := app.LoadProblemSet(name)
		if err != nil {
			return tuiResultMsg{title: "Add a problem set", body: "Failed to load problem set: " + err.Error()}
		}

		summary := app.ImportProblemSet(set)

		var b strings.Builder
		fmt.Fprintf(&b, "%s (%d problems)\n\n", set.Name, len(set.Problems()))
		fmt.Fprintf(&b, "%d added\n", summary.Added)
		fmt.Fprintf(&b, "%d skipped (already present)\n", summary.Skipped)
		fmt.Fprintf(&b, "%d premium placeholders\n", summary.Premium)
		fmt.Fprintf(&b, "%d failed\n", len(summary.Failed))
		for _, problem := range summary.Failed {
			fmt.Fprintf(&b, "\nfailed: %s (%s)", problem.Title, problem.Slug)
		}
		return tuiResultMsg{title: "Add a problem set", body: b.String()}
	}
}

// authenticateCmd saves the cookies and verifies them against LeetCode.
func (app *App) authenticateCmd(session, csrfToken string) tea.Cmd {
	return func() tea.Msg {
		if err := app.SaveAuthentication(session, csrfToken); err != nil {
			return tuiResultMsg{title: "Authenticate user", body: "Failed to save authentication cookies: " + err.Error()}
		}

		user, err := app.fetchUser()
		if err != nil {
			return tuiResultMsg{title: "Authenticate user", body: "Cookies saved, but LeetCode rejected them: " + err.Error()}
		}
		return tuiResultMsg{
			title: "Authenticate user",
			body:  fmt.Sprintf("Signed in as %s (@%s).", user.Data.UserStatus.FullName, user.Data.UserStatus.Username),
		}
	}
}

func (app *App) testProblemCmd(id int) tea.Cmd {
	return func() tea.Msg {
		text, err := app.TestProblem(id)
		if err != nil {
			return tuiResultMsg{title: "Test code", body: "Failed to test code: " + err.Error()}
		}
		return tuiResultMsg{title: "Test code", body: text}
	}
}

func (app *App) submitProblemCmd(id int) tea.Cmd {
	return func() tea.Msg {
		text, err := app.SubmitProblem(id)
		if err != nil {
			return tuiResultMsg{title: "Submit code", body: "Failed to submit code: " + err.Error()}
		}
		return tuiResultMsg{title: "Submit code", body: text}
	}
}
