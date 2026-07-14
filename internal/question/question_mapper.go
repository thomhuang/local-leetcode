package question

import (
	"fmt"
	"strings"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
)

func ToQuestionMap(response AllQuestionsResponse) map[int]QuestionMetadata {
	if len(response.Response) == 0 {
		return map[int]QuestionMetadata{}
	}

	res := make(map[int]QuestionMetadata)
	for _, question := range response.Response {
		res[question.Metadata.FrontEndQuestionId] = question.Metadata
	}

	return res
}

func ToQuestion(response QuestionResponse) Question {
	markdown, _ := htmltomarkdown.ConvertString(response.Data.Question.Content)

	var goCodeSnippet string
	for _, snippet := range response.Data.Question.CodeSnippets {
		if snippet.Language == "Go" {
			titleSlug := response.Data.Question.TitleSlug
			titleSlug = strings.ReplaceAll(titleSlug, "-", "_")
			goCodeSnippet = fmt.Sprintf("package %s\n\n%s", titleSlug, snippet.Code)
			break
		}
	}

	return Question{
		QuestionId:         response.Data.Question.QuestionId,
		FrontEndQuestionId: response.Data.Question.FrontEndQuestionId,
		Title:              response.Data.Question.Title,
		TitleSlug:          response.Data.Question.TitleSlug,
		Content:            markdown,
		Difficulty:         response.Data.Question.Difficulty,
		Language:           "Go",
		CodeSnippet:        goCodeSnippet,
		ExampleTestCases:   response.Data.Question.ExampleTestCases,
	}
}
