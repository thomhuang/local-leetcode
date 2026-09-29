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
			goCodeSnippet = fmt.Sprintf("package %s\n\n%s", PackageName(response.Data.Question.TitleSlug), snippet.Code)
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
		IsPaidOnly:         response.Data.Question.IsPaidOnly,
		Language:           "Go",
		CodeSnippet:        goCodeSnippet,
		ExampleTestCases:   response.Data.Question.ExampleTestCases,
	}
}

// PackageName converts a LeetCode title slug into a valid Go package
// identifier. Go identifiers cannot begin with a digit, so a leading run of
// digits is moved to the end (for example "01-matrix" becomes "matrix01").
func PackageName(titleSlug string) string {
	var sb strings.Builder
	for _, r := range titleSlug {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '_':
			sb.WriteRune(r)
		default:
			sb.WriteByte('_')
		}
	}

	name := sb.String()
	if first := strings.IndexFunc(name, func(r rune) bool { return r < '0' || r > '9' }); first > 0 {
		name = name[first:] + name[:first]
		name = strings.TrimLeft(name, "_")
	}
	if name == "" {
		name = "problem"
	}
	if goKeywords[name] {
		name += "_"
	}
	return name
}

var goKeywords = map[string]bool{
	"break": true, "case": true, "chan": true, "const": true, "continue": true,
	"default": true, "defer": true, "else": true, "fallthrough": true, "for": true,
	"func": true, "go": true, "goto": true, "if": true, "import": true,
	"interface": true, "map": true, "package": true, "range": true, "return": true,
	"select": true, "struct": true, "switch": true, "type": true, "var": true,
}
