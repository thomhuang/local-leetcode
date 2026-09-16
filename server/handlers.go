package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/thomhuang/local-leetcode/internal/question"
	"github.com/thomhuang/local-leetcode/internal/solution"
	"github.com/thomhuang/local-leetcode/internal/user"
)

// decode turns the (body, err) pair from an HTTP call into a typed response.
func decode[T any](data []byte, err error) (T, error) {
	var out T
	if err != nil {
		return out, err
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return out, fmt.Errorf("decode response: %w", err)
	}
	return out, nil
}

func (app *App) fetchQuestion(slug string) (question.Question, error) {
	resp, err := decode[question.QuestionResponse](app.GetQuestion(slug))
	if err != nil {
		return question.Question{}, err
	}
	return question.ToQuestion(resp), nil
}

func (app *App) fetchUser() (user.UserStatusResponse, error) {
	return decode[user.UserStatusResponse](app.GetUser())
}

func (app *App) fetchInterpretation(questionId int, typedCode string) (solution.InterpretSolutionResponse, error) {
	return decode[solution.InterpretSolutionResponse](app.InterpretSolution(questionId, typedCode))
}

func (app *App) fetchCheckResult(interpretId, titleSlug string) (solution.CheckSolutionResponse, error) {
	return decode[solution.CheckSolutionResponse](app.CheckSolution(interpretId, titleSlug))
}

// pollSolution asks LeetCode for the run result until the judge reports a
// final state or pollTimeout passes.
func (app *App) pollSolution(interpretId, titleSlug string) (solution.CheckSolutionResponse, error) {
	if len(interpretId) == 0 {
		return solution.CheckSolutionResponse{}, fmt.Errorf("no interpret id returned; is your session still valid?")
	}

	deadline := time.Now().Add(pollTimeout)
	for {
		resp, err := app.fetchCheckResult(interpretId, titleSlug)
		if err != nil {
			return resp, err
		}
		if solution.IsFinal(resp.State) {
			return resp, nil
		}
		if time.Now().After(deadline) {
			return resp, fmt.Errorf("timed out after %s; last state %q", pollTimeout, resp.State)
		}

		fmt.Println("Pending...")
		time.Sleep(pollInterval)
	}
}
