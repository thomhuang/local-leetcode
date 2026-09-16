package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/thomhuang/local-leetcode/internal/solution"
)

const (
	BaseUrl    = "https://leetcode.com/"
	GraphQlUrl = BaseUrl + "graphql/"
)

func (app *App) GetAllQuestions() ([]byte, error) {
	body, err := app.do(http.MethodGet, BaseUrl+"api/problems/all/", nil, "")
	if err != nil {
		return nil, fmt.Errorf("get all questions: %w", err)
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
		"query": `query questionData($titleSlug: String!) { question(titleSlug: $titleSlug) { questionId questionFrontendId title titleSlug content difficulty likes dislikes exampleTestcases codeSnippets { lang langSlug code } topicTags { name slug } } }`,
	}
	jsonQuery, _ := json.Marshal(query)
	body, err := app.do(http.MethodPost, GraphQlUrl, jsonQuery, "")
	if err != nil {
		return nil, fmt.Errorf("get question %q: %w", slug, err)
	}
	return body, nil
}

func (app *App) InterpretSolution(questionId int, typedCode string) ([]byte, error) {
	ques, err := app.fetchQuestion(app.Questions[questionId].QuestionTitleSlug)
	if err != nil {
		return nil, err
	}

	requestBody := solution.InterpretSolutionRequest{
		DataInput:  ques.ExampleTestCases,
		Language:   "golang",
		QuestionId: strconv.Itoa(app.Questions[questionId].QuestionId),
		TypedCode:  typedCode,
	}
	jsonRequest, _ := json.Marshal(requestBody)
	requestUrl := BaseUrl + "problems/" + ques.TitleSlug + "/interpret_solution/"
	body, err := app.do(http.MethodPost, requestUrl, jsonRequest, BaseUrl+"problems/"+ques.TitleSlug)
	if err != nil {
		return nil, fmt.Errorf("interpret solution: %w", err)
	}
	return body, nil
}

func (app *App) CheckSolution(interpretId, titleSlug string) ([]byte, error) {
	if len(interpretId) == 0 {
		return nil, fmt.Errorf("check solution: interpretId must not be empty")
	}

	requestUrl := BaseUrl + "submissions/detail/" + interpretId + "/check/"
	body, err := app.do(http.MethodGet, requestUrl, nil, BaseUrl+"problems/"+titleSlug)
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
		"query":         `query globalData {  userStatus { realName username } }`,
	}
	jsonQuery, _ := json.Marshal(query)
	body, err := app.do(http.MethodPost, GraphQlUrl, jsonQuery, "")
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	return body, nil
}

// do sends one request to LeetCode and returns the response body. It attaches
// the saved cookies when present and fails on any non-2xx status.
func (app *App) do(method, url string, payload []byte, referer string) ([]byte, error) {
	var body io.Reader
	if payload != nil {
		body = bytes.NewReader(payload)
	}

	request, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	if len(app.UserAuth.AuthCookies) > 0 {
		request.Header.Set("Cookie", app.UserAuth.AuthCookies)
		request.Header.Set("X-Csrftoken", app.UserAuth.CsrfToken)
	}
	if len(referer) > 0 {
		request.Header.Set("Referer", referer)
	}

	resp, err := app.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%s %s returned status %d", method, url, resp.StatusCode)
	}

	return data, nil
}
