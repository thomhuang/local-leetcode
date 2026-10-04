package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/thomhuang/local-leetcode/internal/progress"
	"github.com/thomhuang/local-leetcode/internal/question"
	"github.com/thomhuang/local-leetcode/internal/user"
	"github.com/thomhuang/local-leetcode/util"
)

type App struct {
	Log        *util.Log
	UserAuth   user.UserAuthInfo
	Questions  map[int]question.QuestionMetadata
	Progress   *progress.Store
	client     *http.Client
	chatClient *http.Client
}

func main() {
	importSet := flag.String("import-set", "", "import every problem from the named set under server/output/ProblemSets and exit")
	listSet := flag.String("list", "", "print a checklist for the named problem set and exit")
	category := flag.String("category", "", "with -list, only show problems in this category")
	status := flag.String("status", "", "with -list, only show problems with this status: all, todo, written, accepted, premium, missing")
	flag.Parse()

	if err := os.MkdirAll(outputDir, dirPerm); err != nil {
		fmt.Printf("Unable to create %s: %s\n", outputDir, err)
		os.Exit(1)
	}

	app := NewApp()
	defer app.Log.Close()

	app.log().Info("app.start", "local-leetcode starting",
		"pid", os.Getpid(),
		"args", strings.Join(os.Args[1:], " "),
		"log_file", app.log().Path())

	store, err := progress.Load(progressFile)
	if err != nil {
		app.fail("Unable to load progress", err)
	}
	app.Progress = store

	if *listSet != "" {
		if err := app.printProblemSet(os.Stdout, *listSet, *category, *status); err != nil {
			app.fail("Unable to list problem set", err)
			os.Exit(1)
		}
		return
	}

	app.Questions = app.getQuestions()

	if err := app.ImportAuthentication(); err != nil {
		app.fail("Unable to import existing authentication", err)
	}

	if *importSet != "" {
		app.importSetAndReport(*importSet)
		return
	}

	if err := app.RunREPL(); err != nil {
		app.fail("REPL error", err)
	}
}

func NewApp() *App {
	return &App{
		Log:        util.NewLog(logFile),
		client:     &http.Client{Timeout: httpTimeout},
		chatClient: &http.Client{Timeout: chatTimeout},
	}
}

// log returns the app's logger, or nil when logging is unavailable. The logger
// methods tolerate a nil receiver, so callers can log unconditionally.
func (app *App) log() *util.Log {
	if app == nil {
		return nil
	}
	return app.Log
}

// fail shows an error to the user and records it in the log.
func (app *App) fail(context string, err error) {
	msg := context + ": " + err.Error()
	fmt.Println(msg)
	app.log().Error("app.fail", context, "error", err.Error())
}
