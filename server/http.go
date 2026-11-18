package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/thomhuang/local-leetcode/internal/solution"
)

const (
	BaseUrl    = "https://leetcode.com/"
	GraphQlUrl = BaseUrl + "graphql/"
)

func (app *App) leetcodeBaseURL() string {
	if app != nil && app.baseURL != "" {
		return app.baseURL
	}
	return BaseUrl
}

func (app *App) graphqlURL() string {
	return app.leetcodeBaseURL() + "graphql/"
}

func (app *App) SearchQuestions(keyword string) ([]byte, error) {
	query := map[string]interface{}{
		"operationName": "problemsetQuestionList",
		"variables": map[string]interface{}{
			"categorySlug": "",
			"skip":         0,
			"limit":        50,
			"filters": map[string]interface{}{
				"searchKeywords": keyword,
			},
		},
		"query": `query problemsetQuestionList($categorySlug: String, $limit: Int, $skip: Int, $filters: QuestionListFilterInput) { problemsetQuestionList: questionList(categorySlug: $categorySlug, limit: $limit, skip: $skip, filters: $filters) { total: totalNum questions: data { frontendQuestionId: questionFrontendId title titleSlug } } }`,
	}
	jsonQuery, _ := json.Marshal(query)
	body, err := app.do(http.MethodPost, app.graphqlURL(), jsonQuery, app.leetcodeBaseURL()+"problemset/")
	if err != nil {
		return nil, fmt.Errorf("search questions %q: %w", keyword, err)
	}
	return body, nil
}

func (app *App) GetQuestion(slug string) ([]byte, error) {
	if len(slug) == 0 {
		return nil, fmt.Errorf("get question: slug must not be empty")
	}

	query := map[string]interface{}{
		"operationName": "questionData",
		"variables": map[string]interface{}{
			"titleSlug": slug,
		},
		"query": `query questionData($titleSlug: String!) { question(titleSlug: $titleSlug) { questionId questionFrontendId title titleSlug content difficulty isPaidOnly likes dislikes exampleTestcases codeSnippets { lang langSlug code } topicTags { name slug } } }`,
	}
	jsonQuery, _ := json.Marshal(query)
	body, err := app.do(http.MethodPost, app.graphqlURL(), jsonQuery, app.leetcodeBaseURL()+"problems/"+slug+"/")
	if err != nil {
		return nil, fmt.Errorf("get question %q: %w", slug, err)
	}
	return body, nil
}

func (app *App) InterpretSolution(questionId int, typedCode string) ([]byte, error) {
	meta, err := app.resolveQuestion(questionId)
	if err != nil {
		return nil, err
	}
	ques, err := app.fetchQuestion(meta.QuestionTitleSlug)
	if err != nil {
		return nil, err
	}

	qID := ques.QuestionId
	if qID == "" && meta.QuestionId != 0 {
		qID = strconv.Itoa(meta.QuestionId)
	}
	requestBody := solution.InterpretSolutionRequest{
		DataInput:  ques.ExampleTestCases,
		Language:   "golang",
		QuestionId: qID,
		TypedCode:  typedCode,
	}
	jsonRequest, _ := json.Marshal(requestBody)
	base := app.leetcodeBaseURL()
	requestUrl := base + "problems/" + ques.TitleSlug + "/interpret_solution/"
	body, err := app.do(http.MethodPost, requestUrl, jsonRequest, base+"problems/"+ques.TitleSlug)
	if err != nil {
		return nil, fmt.Errorf("interpret solution: %w", err)
	}
	return body, nil
}

func (app *App) SubmitSolution(questionId int, typedCode string) ([]byte, error) {
	meta, err := app.resolveQuestion(questionId)
	if err != nil {
		return nil, err
	}
	ques, err := app.fetchQuestion(meta.QuestionTitleSlug)
	if err != nil {
		return nil, err
	}

	qID := ques.QuestionId
	if qID == "" && meta.QuestionId != 0 {
		qID = strconv.Itoa(meta.QuestionId)
	}
	requestBody := solution.SubmitSolutionRequest{
		Language:   "golang",
		QuestionId: qID,
		TypedCode:  typedCode,
	}
	jsonRequest, _ := json.Marshal(requestBody)
	base := app.leetcodeBaseURL()
	requestUrl := base + "problems/" + ques.TitleSlug + "/submit/"
	body, err := app.do(http.MethodPost, requestUrl, jsonRequest, base+"problems/"+ques.TitleSlug)
	if err != nil {
		return nil, fmt.Errorf("submit solution: %w", err)
	}
	return body, nil
}

func (app *App) CheckSolution(interpretId, titleSlug string) ([]byte, error) {
	if len(interpretId) == 0 {
		return nil, fmt.Errorf("check solution: interpretId must not be empty")
	}

	base := app.leetcodeBaseURL()
	requestUrl := base + "submissions/detail/" + interpretId + "/check/"
	body, err := app.do(http.MethodGet, requestUrl, nil, base+"problems/"+titleSlug)
	if err != nil {
		return nil, fmt.Errorf("check solution: %w", err)
	}
	return body, nil
}

func (app *App) GetUser() ([]byte, error) {
	if len(app.UserAuth.AuthCookies) == 0 {
		return nil, fmt.Errorf("get user: authenticate first")
	}

	query := map[string]interface{}{
		"operationName": "globalData",
		"variables":     map[string]interface{}{},
		"query":         `query globalData { userStatus { isSignedIn realName username } }`,
	}
	jsonQuery, _ := json.Marshal(query)
	body, err := app.do(http.MethodPost, app.graphqlURL(), jsonQuery, app.leetcodeBaseURL())
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	return body, nil
}

// do sends one request to LeetCode and returns the response body. It attaches
// the saved cookies when present and fails on any non-2xx status.
func (app *App) do(method, url string, payload []byte, referer string) ([]byte, error) {
	started := time.Now()
	var body io.Reader
	if payload != nil {
		body = bytes.NewReader(payload)
	}

	request, err := http.NewRequest(method, url, body)
	if err != nil {
		app.log().Error("http.request", "build request failed", "method", method, "url", url, "error", err.Error())
		return nil, err
	}

	if payload != nil {
		request.Header.Set("Content-Type", "application/json; charset=utf-8")
	}
	if len(app.UserAuth.AuthCookies) > 0 {
		request.Header.Set("Cookie", app.UserAuth.AuthCookies)
		request.Header.Set("X-Csrftoken", app.UserAuth.CsrfToken)
	}
	if len(referer) == 0 {
		referer = app.leetcodeBaseURL()
	}
	request.Header.Set("Referer", referer)

	app.log().Debug("http.request", "sending request",
		"method", method,
		"url", url,
		"authenticated", len(app.UserAuth.AuthCookies) > 0)
	app.log().VerboseDebug("http.request.body", "request payload", "json", string(payload))

	resp, err := app.client.Do(request)
	if err != nil {
		app.log().Error("http.request", "request failed",
			"method", method,
			"url", url,
			"duration_ms", time.Since(started).Milliseconds(),
			"error", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		app.log().Error("http.request", "read response failed",
			"method", method, "url", url, "status", resp.StatusCode, "error", err.Error())
		return nil, fmt.Errorf("read response body: %w", err)
	}

	app.log().Debug("http.response", "response received",
		"method", method,
		"url", url,
		"status", resp.StatusCode,
		"bytes", len(data),
		"duration_ms", time.Since(started).Milliseconds())
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		app.log().VerboseDebug("http.response.body", "error payload", "body", string(data))
		return nil, fmt.Errorf("%s %s returned status %d", method, url, resp.StatusCode)
	}

	return data, nil
}
