package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/thomhuang/local-leetcode/internal/question"
)

func (app *App) getQuestions() map[int]question.QuestionMetadata {
	var stream []byte
	file, err := os.Stat(allProblemsFile)
	// if we can't get the file stats OR it exists, but it's > 5 days old ...
	if err != nil || file.ModTime().Before(time.Now().AddDate(0, 0, -5)) {
		stream, err = app.GetAllQuestions()
		if err != nil {
			app.Log.Append("File doesn't exist or stale!")
			return map[int]question.QuestionMetadata{}
		}

		err := os.WriteFile(allProblemsFile, stream, os.ModePerm)
		if err != nil {
			app.Log.Append(fmt.Sprintf("Unable to cache problems json, %s", err.Error()))
		}
	} else {
		stream, err = os.ReadFile(allProblemsFile)
		if err != nil {
			app.Log.Append(fmt.Sprintf("failed to read cached file %s", err.Error()))
			return map[int]question.QuestionMetadata{}
		}
	}

	var questions question.AllQuestionsResponse
	err = json.Unmarshal(stream, &questions)
	if err != nil {
		app.Log.Append(fmt.Sprintf("could not unmarshal problems metadata response body: %s\n", err.Error()))
		return map[int]question.QuestionMetadata{}
	}
	if len(questions.Response) == 0 {
		app.Log.Append("No problems found with no error!")
		return map[int]question.QuestionMetadata{}
	}

	response := question.ToQuestionMap(questions)
	app.Log.Append(fmt.Sprintf("getQuestions response: %+v", response))

	return response
}
