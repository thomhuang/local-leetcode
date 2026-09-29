package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/thomhuang/local-leetcode/internal/question"
	"github.com/thomhuang/local-leetcode/internal/user"
	"github.com/thomhuang/local-leetcode/util"
)

type App struct {
	Log       *util.Log
	UserAuth  user.UserAuthInfo
	Questions map[int]question.QuestionMetadata
	client    *http.Client
}

func main() {
	importSet := flag.String("import-set", "", "import every problem from the named set under server/output/ProblemSets and exit")
	flag.Parse()

	if err := os.MkdirAll(outputDir, dirPerm); err != nil {
		fmt.Printf("Unable to create %s: %s\n", outputDir, err)
		os.Exit(1)
	}

	app := NewApp()
	defer app.Log.Close()

	app.Questions = app.getQuestions()

	err := app.ImportAuthentication()
	if err != nil {
		app.fail("Unable to import existing authentication", err)
	}

	if *importSet != "" {
		app.importSetAndReport(*importSet)
		return
	}

	app.Prompt()
}

func NewApp() *App {
	return &App{
		Log:    util.NewLog(logFile),
		client: &http.Client{Timeout: httpTimeout},
	}
}

// fail shows an error to the user and records it in the log.
func (app *App) fail(context string, err error) {
	msg := context + ": " + err.Error()
	fmt.Println(msg)
	app.Log.Append(msg)
}
