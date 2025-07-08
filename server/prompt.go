package main

import (
	"bufio"
	"fmt"
	"github.com/thomhuang/local-leetcode/util"
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

func (server *HttpServer) Prompt() {
	scanner := bufio.NewScanner(os.Stdin)
	var exit bool
	for !exit {
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
		PromptWithValidation(
			scanner,
			sb.String(),
			func(input string) (bool, string) {
				valid := true
				var msg string
				switch input {
				case "1":
					action = AddQuestion
					break
				case "2":
					action = Authenticate
					break
				case "3":
					action = TestCode
					break
				case "4":
					action = SubmitCode
					break
				case "5":
					action = Exit
				default:
					action = Invalid
					valid, msg = false, "Please input a valid action!"
				}
				return valid, msg
			})

		switch action {
		case AddQuestion:
			var id int
			PromptWithValidation(
				scanner,
				"What problem number are you interested in?",
				func(input string) (bool, string) {
					convertedId, err := strconv.Atoi(input)
					if err != nil {
						return false, fmt.Sprintf("Invalid input for problem number, please input an integer!: %s", input)
					}

					id = convertedId
					return true, ""
				},
			)

			err := util.SaveMarkdownContent(getQuestion(server.Questions[id]))
			if err != nil {
				server.Log.Append(fmt.Sprintf("Failed to save question content! %s", err.Error()))
				return
			}
			break
		case Authenticate:
			lastUpdated := server.UserAuth.LastUpdated
			// We should be authenticated for 7 days, so if already authenticated within 5, ask if they still want to ...
			fiveDaysPrev := time.Now().AddDate(0, 0, -5)
			if fiveDaysPrev.Before(lastUpdated) {
				daysDiff := int(lastUpdated.Sub(fiveDaysPrev).Hours() / 24)
				msg := fmt.Sprintf("Are you sure you still want to authenticate? You have %d days left!", daysDiff)
				doAuth := promptYesNo(scanner, msg)
				if !doAuth {
					break
				}
			}

			cookies := make(map[string]string)
			PromptWithValidation(
				scanner,
				"Please input your authenticated request cookies from a https://leetcode.com/graphql call!",
				func(input string) (bool, string) {
					rawCookies := input
					pairs := strings.Split(rawCookies, ";")

					for _, pair := range pairs {
						curr := strings.Split(strings.TrimSpace(pair), "=")
						if curr[0] == "csrftoken" || curr[0] == "LEETCODE_SESSION" {
							cookies[curr[0]] = curr[1]
						}
					}

					if len(cookies) != 2 {
						return false, "Invalid cookie value, please try again!"
					}
					return true, ""
				},
			)

			err := server.SaveAuthentication(cookies)
			if err != nil {
				server.Log.Append(fmt.Sprintf("Failed to save authentication cookies! %s", err.Error()))
			}

			user := GetUser()

			fmt.Println(user.Data.UserStatus.FullName)
			fmt.Println(user.Data.UserStatus.Username)
		case Exit:
			fmt.Println("Exiting ...")
			exit = true
			break
		default:
			panic("unhandled default case")
		}
	}

	server.Log.OutputLogFile()
}

func promptYesNo(scanner *bufio.Scanner, question string) bool {
	for {
		fmt.Printf("%s Y/N\n", question)
		input := readInput(scanner)

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

func PromptWithValidation(scanner *bufio.Scanner, prompt string, validate func(string) (bool, string)) string {
	for {
		fmt.Println(prompt)
		input := readInput(scanner)
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

func readInput(scanner *bufio.Scanner) string {
	scanner.Scan()
	return scanner.Text()
}
