package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	q "github.com/thomhuang/local-leetcode/internal/question"
	s "github.com/thomhuang/local-leetcode/internal/solution"
	u "github.com/thomhuang/local-leetcode/internal/user"
)

func (server *HttpServer) getQuestions() map[int]string {
	var stream []byte
	file, err := os.Stat("./output/all_problems.json")
	// if we can't get the file stats OR it exists, but it's > 5 days old ...
	if err != nil || file.ModTime().Before(time.Now().AddDate(0, 0, -5)) {
		stream, err = server.GetAllQuestions()
		if err != nil {
			return map[int]string{}
		}

		err := os.WriteFile("./output/all_problems.json", stream, os.ModePerm)
		if err != nil {
			server.Log.Append(fmt.Sprintf("Unable to cache problems json, %s", err.Error()))
		}
	} else {
		stream, err = os.ReadFile("./output/all_problems.json")
		if err != nil {
			server.Log.Append(fmt.Sprintf("failed to read cached file %s", err.Error()))
			return map[int]string{}
		}
	}

	var questions q.AllQuestionsResponse
	err = json.Unmarshal(stream, &questions)
	if err != nil {
		server.Log.Append(fmt.Sprintf("could not unmarshal problems metadata response body: %s\n", err.Error()))
		return map[int]string{}
	}
	if len(questions.Response) == 0 {
		server.Log.Append("No problems found with no error!")
		return map[int]string{}
	}

	return q.ToQuestionMap(questions)
}

func getQuestion(slug string) q.Question {
	stream, err := server.GetQuestion(slug)
	if err != nil {
		return q.Question{}
	}

	var question q.QuestionResponse
	err = json.Unmarshal(stream, &question)
	if err != nil {
		return q.Question{}
	}

	return q.ToQuestion(question)
}

func GetUser() u.UserStatusResponse {
	stream, err := server.GetUser()
	if err != nil {
		return u.UserStatusResponse{}
	}

	var user u.UserStatusResponse
	err = json.Unmarshal(stream, &user)
	if err != nil {
		return u.UserStatusResponse{}
	}

	return user
}

func InterpretSolution(questionId int, typedCode string) s.InterpretSolutionResponse {
	stream, err := server.InterpretSolution(questionId, typedCode)
	if err != nil {
		return s.InterpretSolutionResponse{}
	}

	var interpretation s.InterpretSolutionResponse
	err = json.Unmarshal(stream, &interpretation)
	if err != nil {
		return s.InterpretSolutionResponse{}
	}

	return interpretation
}

func CheckSolution(interpretId, titleSlug string) s.CheckSolutionResponse {
	stream, err := server.CheckSolution(interpretId, titleSlug)
	if err != nil {
		return s.CheckSolutionResponse{}
	}

	var submissionStatus s.CheckSolutionResponse
	err = json.Unmarshal(stream, &submissionStatus)
	if err != nil {
		return s.CheckSolutionResponse{}
	}

	return submissionStatus
}

func PollCheckSolution(interpretId, titleSlug string) s.CheckSolutionResponse {
	if len(interpretId) == 0 {
		return s.CheckSolutionResponse{}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ticker := time.NewTicker(time.Second) // Poll every 1s
	defer ticker.Stop()

	var resp s.CheckSolutionResponse
	for {
		select {
		case <-ctx.Done():
			return resp
		case <-ticker.C:
			if resp.State == Started {
				// Final call to get the data
				fmt.Println("Submitted!")
				return CheckSolution(interpretId, titleSlug)
			} else {
				fmt.Println("Pending...")
				resp = CheckSolution(interpretId, titleSlug)
			}
		}
	}
}
