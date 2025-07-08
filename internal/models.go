package internal

import "time"

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
	Data QuestionsModel `json:"data"`
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
	QuestionId  string
	Title       string
	TitleSlug   string
	Content     string
	Difficulty  string
	Language    string
	CodeSnippet string
}

type UserAuthInfo struct {
	AuthCookies string
	LastUpdated time.Time
}

type UserStatusResponse struct {
	Data UserStatusModel `json:"data"`
}

type UserStatusModel struct {
	UserStatus UserModel `json:"userStatus"`
}

type UserModel struct {
	FullName string `json:"realName"`
	Username string `json:"username"`
}
