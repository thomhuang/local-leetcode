package main

import (
	"os"

	"github.com/thomhuang/local-leetcode/internal/question"
	"github.com/thomhuang/local-leetcode/internal/user"
	"github.com/thomhuang/local-leetcode/util"
)

type App struct {
	Log       *util.Log
	UserAuth  user.UserAuthInfo
	Questions map[int]question.QuestionMetadata
}

func main() {
	_ = os.MkdirAll(outputDir, os.ModeDir)

	app := NewApp()

	app.Questions = app.getQuestions()

	err := app.ImportAuthentication()
	if err != nil {
		app.Log.Append("Unable to import existing authentication! " + err.Error())
	}

	app.Prompt()
}

func NewApp() *App {
	return &App{
		Log: util.NewLog(),
	}
}
