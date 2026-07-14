package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"
)

type UserAction int

const (
	Invalid UserAction = iota
	Authenticate
	AddQuestion
	TestCode
	SubmitCode
	Exit
)

func (app *App) Prompt() {
	scanner := bufio.NewScanner(os.Stdin)
	var exit bool
	for !exit {
		action := app.promptForAction(scanner)

		switch action {
		case AddQuestion:
			app.handleAddQuestion(scanner)
		case Authenticate:
			app.handleAuthentication(scanner)
		case TestCode:
			app.handleTestCode(scanner)
		case SubmitCode:
			// TODO: Implement submit code functionality
			fmt.Println("Submit code functionality not yet implemented")
		case Exit:
			fmt.Println("Exiting ...")
			exit = true
		default:
			panic("unhandled default case")
		}
	}

	app.Log.OutputLogFile()
}

func (app *App) promptForAction(scanner *bufio.Scanner) UserAction {
	var action UserAction
	var sb strings.Builder
	sb.WriteString("--------------------------------------------------------------------\n")
	sb.WriteString("Please select your action of choice (only expecting the number):\n")
	sb.WriteString("(1) Add a new question\n")
	sb.WriteString("(2) Authenticate user\n")
	sb.WriteString("(3) Test code\n")
	sb.WriteString("(4) Submit code\n")
	sb.WriteString("(5) Exit\n")
	sb.WriteString("--------------------------------------------------------------------")

	app.promptWithValidation(
		scanner,
		sb.String(),
		func(input string) (bool, string) {
			valid := true
			var msg string
			switch input {
			case "1":
				action = AddQuestion
			case "2":
				action = Authenticate
			case "3":
				action = TestCode
			case "4":
				action = SubmitCode
			case "5":
				action = Exit
			default:
				action = Invalid
				valid, msg = false, "Please input a valid action!"
			}
			return valid, msg
		})

	return action
}

func (app *App) handleAddQuestion(scanner *bufio.Scanner) {
	id := app.promptForQuestionID(scanner, "What problem number are you interested in?")

	err := SaveMarkdownContent(app.fetchQuestion(app.Questions[id].QuestionTitleSlug))
	if err != nil {
		app.Log.Append("Failed to save question content! " + err.Error())
		return
	}
}

func (app *App) handleAuthentication(scanner *bufio.Scanner) {
	if app.shouldSkipAuthentication(scanner) {
		return
	}

	cookies := app.promptForCookies(scanner)
	err := app.SaveAuthentication(cookies)
	if err != nil {
		app.Log.Append("Failed to save authentication cookies! " + err.Error())
	}

	user := app.fetchUser()
	fmt.Println(user.Data.UserStatus.FullName)
	fmt.Println(user.Data.UserStatus.Username)
}

func (app *App) handleTestCode(scanner *bufio.Scanner) {
	id := app.promptForQuestionID(scanner, "What problem number would you like to run? Please make sure it exists under /output/problems/{titleSlug}")

	titleSlug := app.Questions[id].QuestionTitleSlug
	filePath := problemsDir + "/" + titleSlug + "/" + strconv.Itoa(id) + "-" + titleSlug + ".go"

	fileStream, err := os.ReadFile(filePath)
	if err != nil {
		app.Log.Append("Failed to read problem file! " + err.Error())
		return
	}

	packageName := strings.ReplaceAll(titleSlug, "-", "_")
	userSubmission := strings.ReplaceAll(
		string(fileStream),
		"package "+packageName+"\n\n",
		"")

	pendingSolution := app.fetchInterpretation(id, userSubmission)
	interpretedId := pendingSolution.InterpretId

	result := app.pollSolution(interpretedId, titleSlug)
	fmt.Println(OutputQuestionResults(result))
}

func (app *App) promptForQuestionID(scanner *bufio.Scanner, prompt string) int {
	var id int
	app.promptWithValidation(
		scanner,
		prompt,
		func(input string) (bool, string) {
			convertedId, err := strconv.Atoi(input)
			if err != nil {
				return false, fmt.Sprintf("Invalid input for problem number, please input an integer!: %s", input)
			}

			id = convertedId
			return true, ""
		},
	)
	return id
}

func (app *App) shouldSkipAuthentication(scanner *bufio.Scanner) bool {
	lastUpdated := app.UserAuth.LastUpdated
	fiveDaysPrev := time.Now().AddDate(0, 0, -authFreshnessDays)
	if fiveDaysPrev.Before(lastUpdated) {
		daysDiff := int(lastUpdated.Sub(fiveDaysPrev).Hours() / 24)
		msg := fmt.Sprintf("Are you sure you still want to authenticate? You have %d days left!", daysDiff)
		return !app.promptYesNo(scanner, msg)
	}
	return false
}

func (app *App) promptForCookies(scanner *bufio.Scanner) map[string]string {
	cookies := make(map[string]string)
	app.promptWithValidation(
		scanner,
		"Please input your authenticated request cookies from a https://leetcode.com/graphql call!",
		func(input string) (bool, string) {
			rawCookies := input
			pairs := strings.Split(rawCookies, ";")

			for _, pair := range pairs {
				curr := strings.Split(strings.TrimSpace(pair), "=")
				if curr[0] == "csrftoken" {
					cookies[curr[0]] = curr[1]
					app.UserAuth.CsrfToken = curr[1]
				}
				if curr[0] == "LEETCODE_SESSION" {
					cookies[curr[0]] = curr[1]
				}
			}

			if len(cookies) != 2 {
				return false, "Invalid cookie value, please try again!"
			}
			return true, ""
		},
	)
	return cookies
}

func (app *App) promptYesNo(scanner *bufio.Scanner, question string) bool {
	for {
		fmt.Printf("%s Y/N\n", question)
		input := app.readInput(scanner)

		if len(input) != 1 {
			fmt.Println("Please input Y or N!")
			continue
		}

		switch unicode.ToLower(rune(input[0])) {
		case 'y':
			fmt.Println()
			return true
		case 'n':
			fmt.Println()
			return false
		default:
			fmt.Println("Please input Y or N!")
		}
	}
}

func (app *App) promptWithValidation(scanner *bufio.Scanner, prompt string, validate func(string) (bool, string)) string {
	for {
		fmt.Println(prompt)
		input := app.readInput(scanner)
		fmt.Println()

		if valid, errorMsg := validate(input); valid {
			return input
		} else {
			if len(errorMsg) > 0 {
				fmt.Println(errorMsg)
			}
			fmt.Println()
		}
	}
}

func (app *App) readInput(scanner *bufio.Scanner) string {
	scanner.Scan()
	return scanner.Text()
}
