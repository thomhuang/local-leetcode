package main

import (
	"encoding/json"
	"fmt"
	"github.com/thomhuang/local-leetcode/internal"
	"os"
	"time"
)

func (server *HttpServer) getQuestions() map[int]string {
	var questionsStream []byte
	file, err := os.Stat("./output/all_problems.json")
	// if we can't get the file stats OR it exists, but it's > 5 days old ...
	if err != nil || file.ModTime().Before(time.Now().AddDate(0, 0, -5)) {
		questionsStream, err = server.GetAllQuestions()
		if err != nil {
			return map[int]string{}
		}

		err := os.WriteFile("./output/all_problems.json", questionsStream, os.ModePerm)
		if err != nil {
			server.Log.Append(fmt.Sprintf("Unable to cache problems json, %s", err.Error()))
		}
	} else {
		questionsStream, err = os.ReadFile("./output/all_problems.json")
		if err != nil {
			server.Log.Append(fmt.Sprintf("failed to read cached file %s", err.Error()))
			return map[int]string{}
		}
	}

	var questions internal.AllQuestionsResponse
	err = json.Unmarshal(questionsStream, &questions)
	if err != nil {
		server.Log.Append(fmt.Sprintf("could not unmarshal problems metadata response body: %s\n", err.Error()))
		return map[int]string{}
	}
	if len(questions.Response) == 0 {
		server.Log.Append("No problems found with no error!")
		return map[int]string{}
	}

	return internal.ToQuestionMap(questions)
}

func getQuestion(slug string) internal.Question {
	questionStream, err := server.GetQuestion(slug)
	if err != nil {
		return internal.Question{}
	}

	var question internal.QuestionResponse
	err = json.Unmarshal(questionStream, &question)
	if err != nil {
		return internal.Question{}
	}

	return internal.ToQuestion(question)
}

func GetUser() internal.UserStatusResponse {
	userInfoStream, err := server.GetUser()
	if err != nil {
		return internal.UserStatusResponse{}
	}

	var user internal.UserStatusResponse
	err = json.Unmarshal(userInfoStream, &user)
	if err != nil {
		return internal.UserStatusResponse{}
	}

	return user
}
