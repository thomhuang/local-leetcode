package src

type AllQuestionsResponse struct {
	Response []QuestionsMetadataModel `json:"stat_status_pairs"`
}

type QuestionsMetadataModel struct {
	Metadata QuestionMetadataModel `json:"stat"`
}

type QuestionMetadataModel struct {
	QuestionId        int    `json:"question_id"`
	QuestionTitleSlug string `json:"question__title_slug"`
}

type QuestionResponse struct {
	Response QuestionsModel `json:"data"`
}

type QuestionsModel struct {
	QuestionInfo QuestionModel `json:"question"`
}

type QuestionModel struct {
	QuestionId       string        `json:"questionId"`
	Title            string        `json:"title"`
	TitleSlug        string        `json:"titleSlug"`
	Content          string        `json:"content"`
	Difficulty       string        `json:"difficulty"`
	ExampleTestCases string        `json:"exampleTestCases"`
	CodeSnippets     []CodeSnippet `json:"codeSnippets"`
}

type CodeSnippet struct {
	Language     string `json:"lang"`
	LanguageSlug string `json:"langSlug"`
	Code         string `json:"code"`
}

type Question struct {
	Title       string
	Content     string
	Difficulty  string
	Language    string
	CodeSnippet string
}
