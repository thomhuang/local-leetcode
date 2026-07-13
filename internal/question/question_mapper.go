package question

import (
	"fmt"
	"strings"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
)

func ToQuestionMap(response AllQuestionsResponse) map[int]QuestionMetadataModel {
	if len(response.Response) == 0 {
		return map[int]QuestionMetadataModel{}
	}

	res := make(map[int]QuestionMetadataModel)
	for _, question := range response.Response {
		res[question.Metadata.FrontEndQuestionId] = question.Metadata
	}

	return res
}

func ToQuestion(response QuestionResponse) Question {
	markdown, _ := htmltomarkdown.ConvertString(response.Data.QuestionInfo.Content)

	var goCodeSnippet string
	for _, snippet := range response.Data.QuestionInfo.CodeSnippets {
		if snippet.Language == "Go" {
			titleSlug := response.Data.QuestionInfo.TitleSlug
			titleSlug = strings.ReplaceAll(titleSlug, "-", "_")
			goCodeSnippet = fmt.Sprintf("package %s\n\n%s", titleSlug, snippet.Code)
			break
		}
	}

	return Question{
		QuestionId:         response.Data.QuestionInfo.QuestionId,
		FrontEndQuestionId: response.Data.QuestionInfo.FrontEndQuestionId,
		Title:              response.Data.QuestionInfo.Title,
		TitleSlug:          response.Data.QuestionInfo.TitleSlug,
		Content:            markdown,
		Difficulty:         response.Data.QuestionInfo.Difficulty,
		Language:           "Go",
		CodeSnippet:        goCodeSnippet,
		ExampleTestCases:   response.Data.QuestionInfo.ExampleTestCases,
	}
}
