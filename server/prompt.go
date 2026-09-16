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

	ques, err := app.fetchQuestion(app.Questions[id].QuestionTitleSlug)
	if err != nil {
		app.fail("Failed to fetch question", err)
		return
	}

	err = SaveMarkdownContent(ques)
	if err != nil {
		app.fail("Failed to save question content", err)
		return
	}
}

func (app *App) handleAuthentication(scanner *bufio.Scanner) {
	if app.shouldSkipAuthentication(scanner) {
		return
	}

	session, csrfToken := app.promptForCookies(scanner)
	err := app.SaveAuthentication(session, csrfToken)
	if err != nil {
		app.fail("Failed to save authentication cookies", err)
		return
	}

	user, err := app.fetchUser()
	if err != nil {
		app.fail("Cookies saved, but LeetCode rejected them", err)
		return
	}
	fmt.Println(user.Data.UserStatus.FullName)
	fmt.Println(user.Data.UserStatus.Username)
}

func (app *App) handleTestCode(scanner *bufio.Scanner) {
	id := app.promptForQuestionID(scanner, "What problem number would you like to run? Please make sure it exists under /output/problems/{titleSlug}")

	titleSlug := app.Questions[id].QuestionTitleSlug
	filePath := problemsDir + "/" + titleSlug + "/" + strconv.Itoa(id) + "-" + titleSlug + ".go"

	fileStream, err := os.ReadFile(filePath)
	if err != nil {
		app.fail("Failed to read problem file", err)
		return
	}

	packageName := strings.ReplaceAll(titleSlug, "-", "_")
	userSubmission := prepareSubmission(string(fileStream), packageName)

	pendingSolution, err := app.fetchInterpretation(id, userSubmission)
	if err != nil {
		app.fail("Failed to submit code for a test run", err)
		return
	}

	result, err := app.pollSolution(pendingSolution.InterpretId, titleSlug)
	if err != nil {
		app.fail("Failed to get the run result", err)
		return
	}
	fmt.Println(OutputQuestionResults(result))
}

func prepareSubmission(source, packageName string) string {
	packageLine, code, found := strings.Cut(source, "\n")
	packageLine = strings.TrimSpace(strings.TrimPrefix(packageLine, "\uFEFF"))
	if !found || packageLine != "package "+packageName {
		return source
	}

	return strings.TrimLeft(code, "\r\n")
}

// parseCookies pulls the session and CSRF token out of a raw Cookie header.
// The last value is false when either cookie is missing.
func parseCookies(raw string) (session, csrfToken string, ok bool) {
	for _, pair := range strings.Split(raw, ";") {
		name, value, found := strings.Cut(strings.TrimSpace(pair), "=")
		if !found {
			continue
		}
		switch name {
		case "LEETCODE_SESSION":
			session = value
		case "csrftoken":
			csrfToken = value
		}
	}
	return session, csrfToken, session != "" && csrfToken != ""
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
			if _, exists := app.Questions[convertedId]; !exists {
				return false, fmt.Sprintf("Problem %d was not found in the problem list!", convertedId)
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

func (app *App) promptForCookies(scanner *bufio.Scanner) (session, csrfToken string) {
	app.promptWithValidation(
		scanner,
		"Please input your authenticated request cookies from a https://leetcode.com/graphql call!",
		func(input string) (bool, string) {
			var ok bool
			session, csrfToken, ok = parseCookies(input)
			if !ok {
				return false, "Cookies must contain LEETCODE_SESSION and csrftoken, please try again!"
			}
			return true, ""
		},
	)
	return session, csrfToken
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
