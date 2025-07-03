package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/thomhuang/local-leetcode/src"
	"github.com/thomhuang/local-leetcode/util"
)

var server *src.HttpServer
var questions map[int]string

func main() {
	_ = os.MkdirAll("output", os.ModePerm)
	scanner := bufio.NewScanner(os.Stdin)
	server = src.NewHttpServer()

	questions = getQuestions()

	var id int
	util.PromptWithValidation(
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

	question := getQuestion(questions[id])

	fmt.Println(question)
	server.Log.OutputLogFile()
}

func getQuestions() map[int]string {
	questionsStream, err := os.ReadFile("./output/all_problems.json")
	if err != nil {
		server.Log.Append(fmt.Sprintf("Cached file doesn't exist or we're unable to read cached file, %s", err.Error()))
		questionsStream, err = server.GetAllQuestions()
		if err != nil {
			server.Log.Append(fmt.Sprintf("failed to fetch all problems from leetcode, %s", err.Error()))
			return map[int]string{}
		}

		err := os.WriteFile("./output/all_problems.json", questionsStream, 0644)
		if err != nil {
			server.Log.Append(fmt.Sprintf("Unable to cache problems json, %s", err.Error()))
		}
	}

	var questions src.AllQuestionsResponse
	err = json.Unmarshal(questionsStream, &questions)
	if err != nil {
		server.Log.Append(fmt.Sprintf("could not unmarshal problems metadata response body: %s\n", err.Error()))
		return map[int]string{}
	}
	if len(questions.Response) == 0 {
		server.Log.Append("No problems found with no error!")
		return map[int]string{}
	}

	return src.ToQuestionMap(questions)
}

func getQuestion(slug string) src.Question {
	questionStream, err := server.GetQuestion(slug)
	if err != nil {
		server.Log.Append(fmt.Sprintf("failed to fetch problem info from leetcode. Problem: %s. %s", slug, err.Error()))
		return src.Question{}
	}

	var question src.QuestionResponse
	err = json.Unmarshal(questionStream, &question)
	if err != nil {
		server.Log.Append(fmt.Sprintf("could not unmarshal problems metadata response body: %s\n", err.Error()))
		return src.Question{}
	}

	return src.ToQuestion(question)
}
