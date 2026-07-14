package question

type AllQuestionsResponse struct {
	Response []QuestionStatPair `json:"stat_status_pairs"`
}

type QuestionStatPair struct {
	Metadata QuestionMetadata `json:"stat"`
}

type QuestionMetadata struct {
	QuestionId         int    `json:"question_id"`
	QuestionTitleSlug  string `json:"question__title_slug"`
	FrontEndQuestionId int    `json:"frontend_question_id"`
}

type QuestionResponse struct {
	Data QuestionData `json:"data"`
}

type QuestionData struct {
	Question QuestionInfo `json:"question"`
}

type QuestionInfo struct {
	QuestionId         string        `json:"questionId"`
	FrontEndQuestionId string        `json:"questionFrontendId"`
	Title              string        `json:"title"`
	TitleSlug          string        `json:"titleSlug"`
	Content            string        `json:"content"`
	Difficulty         string        `json:"difficulty"`
	ExampleTestCases   string        `json:"exampleTestcases"`
	CodeSnippets       []CodeSnippet `json:"codeSnippets"`
}

type CodeSnippet struct {
	Language     string `json:"lang"`
	LanguageSlug string `json:"langSlug"`
	Code         string `json:"code"`
}

type Question struct {
	QuestionId         string
	FrontEndQuestionId string
	Title              string
	TitleSlug          string
	Content            string
	Difficulty         string
	Language           string
	CodeSnippet        string
	ExampleTestCases   string
}
