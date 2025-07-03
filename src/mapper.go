package src

import (
	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
)

func ToQuestionMap(response AllQuestionsResponse) map[int]string {
	if len(response.Response) == 0 {
		return map[int]string{}
	}

	res := make(map[int]string)
	for _, question := range response.Response {
		res[question.Metadata.QuestionId] = question.Metadata.QuestionTitleSlug
	}

	return res
}

func ToQuestion(response QuestionResponse) Question {
	markdown, _ := htmltomarkdown.ConvertString(response.Response.QuestionInfo.Content)

	var goCodeSnippet string
	for _, snippet := range response.Response.QuestionInfo.CodeSnippets {
		if snippet.Language == "Go" {
			goCodeSnippet = snippet.Code
		}
	}

	return Question{
		Title:       response.Response.QuestionInfo.Title,
		Content:     markdown,
		Difficulty:  response.Response.QuestionInfo.Difficulty,
		Language:    "Go",
		CodeSnippet: goCodeSnippet,
	}
}
