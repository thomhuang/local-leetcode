package main

import (
	"encoding/json"
	"os"
	"time"

	"github.com/thomhuang/local-leetcode/internal/question"
)

func (app *App) getQuestions() map[int]question.QuestionMetadata {
	var stream []byte
	file, err := os.Stat(allProblemsFile)
	// if we can't get the file stats OR it exists, but it's > 5 days old ...
	if err != nil || file.ModTime().Before(time.Now().AddDate(0, 0, -cacheFreshnessDays)) {
		app.log().Info("questions.cache", "refreshing the problem list", "path", allProblemsFile)
		stream, err = app.GetAllQuestions()
		if err != nil {
			app.fail("Unable to download the problem list", err)
			return map[int]question.QuestionMetadata{}
		}

		err := os.WriteFile(allProblemsFile, stream, filePerm)
		if err != nil {
			app.log().Warn("questions.cache", "unable to cache problem list", "path", allProblemsFile, "error", err.Error())
		}
	} else {
		app.log().Debug("questions.cache", "using cached problem list", "path", allProblemsFile, "modified", file.ModTime().Format(time.RFC3339))
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
		app.log().Warn("questions.cache", "problem list is empty but parsed without error")
		return map[int]question.QuestionMetadata{}
	}

	response := question.ToQuestionMap(questions)
	app.log().Info("questions.loaded", "problem list ready", "problems", len(response))

	return response
}
