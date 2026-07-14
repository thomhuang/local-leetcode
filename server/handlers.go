package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/thomhuang/local-leetcode/internal/question"
	"github.com/thomhuang/local-leetcode/internal/solution"
	"github.com/thomhuang/local-leetcode/internal/user"
)

func (app *App) fetchQuestion(slug string) question.Question {
	stream, err := app.GetQuestion(slug)
	if err != nil {
		return question.Question{}
	}

	var resp question.QuestionResponse
	err = json.Unmarshal(stream, &resp)
	if err != nil {
		return question.Question{}
	}

	return question.ToQuestion(resp)
}

func (app *App) fetchUser() user.UserStatusResponse {
	stream, err := app.GetUser()
	if err != nil {
		return user.UserStatusResponse{}
	}

	var u user.UserStatusResponse
	err = json.Unmarshal(stream, &u)
	if err != nil {
		return user.UserStatusResponse{}
	}

	return u
}

func (app *App) fetchInterpretation(questionId int, typedCode string) solution.InterpretSolutionResponse {
	stream, err := app.InterpretSolution(questionId, typedCode)
	if err != nil {
		return solution.InterpretSolutionResponse{}
	}

	var interpretation solution.InterpretSolutionResponse
	err = json.Unmarshal(stream, &interpretation)
	if err != nil {
		return solution.InterpretSolutionResponse{}
	}

	return interpretation
}

func (app *App) fetchCheckResult(interpretId, titleSlug string) solution.CheckSolutionResponse {
	stream, err := app.CheckSolution(interpretId, titleSlug)
	if err != nil {
		return solution.CheckSolutionResponse{}
	}

	var submissionStatus solution.CheckSolutionResponse
	err = json.Unmarshal(stream, &submissionStatus)
	if err != nil {
		return solution.CheckSolutionResponse{}
	}

	return submissionStatus
}

func (app *App) pollSolution(interpretId, titleSlug string) solution.CheckSolutionResponse {
	if len(interpretId) == 0 {
		return solution.CheckSolutionResponse{}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ticker := time.NewTicker(time.Second) // Poll every 1s
	defer ticker.Stop()

	var resp solution.CheckSolutionResponse
	for {
		select {
		case <-ctx.Done():
			return resp
		case <-ticker.C:
			if resp.State == solution.Started {
				// Final call to get the data
				fmt.Println("Submitted!")
				return app.fetchCheckResult(interpretId, titleSlug)
			} else {
				fmt.Println("Pending...")
				resp = app.fetchCheckResult(interpretId, titleSlug)
			}
		}
	}
}
