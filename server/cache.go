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
	if err != nil || file.ModTime().Before(time.Now().AddDate(0, 0, -cacheFreshnessDays)) {
		stream, err = app.GetAllQuestions()
		if err != nil {
			app.fail("Unable to download the problem list", err)
			return map[int]question.QuestionMetadata{}
		}

		err := os.WriteFile(allProblemsFile, stream, filePerm)
		if err != nil {
			app.Log.Append(fmt.Sprintf("Unable to cache problems json, %s", err.Error()))
		}
	} else {
		stream, err = os.ReadFile(allProblemsFile)
		if err != nil {
			app.fail("Unable to read the cached problem list", err)
			return map[int]question.QuestionMetadata{}
		}
	}

	var questions question.AllQuestionsResponse
	err = json.Unmarshal(stream, &questions)
	if err != nil {
		app.fail("Unable to parse the problem list", err)
		return map[int]question.QuestionMetadata{}
	}
	if len(questions.Response) == 0 {
		app.Log.Append("No problems found with no error!")
		return map[int]question.QuestionMetadata{}
	}

	response := question.ToQuestionMap(questions)
	app.Log.Append(fmt.Sprintf("getQuestions: loaded %d problems", len(response)))

	return response
}
