package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/thomhuang/local-leetcode/internal/question"
	"github.com/thomhuang/local-leetcode/internal/solution"
)

const (
	BaseUrl    = "https://leetcode.com/"
	GraphQlUrl = BaseUrl + "graphql/"
)

func (app *App) GetAllQuestions() ([]byte, error) {
	resp, err := http.Get(BaseUrl + "api/problems/all/")
	if err != nil {
		app.Log.Append(fmt.Sprintf("GetAllQuestions: could not download problems metadata: %s\n", err.Error()))
		return []byte{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		app.Log.Append(fmt.Sprintf("GetAllQuestions: could not read problems metadata response body: %s\n", err.Error()))
		return []byte{}, err
	}

	return body, nil
}

func (app *App) GetQuestion(slug string) ([]byte, error) {
	if len(slug) <= 0 {
		app.Log.Append("GetQuestion: title slug must not be empty")
		return []byte{}, fmt.Errorf("slug must not be empty")
	}

	query := map[string]interface{}{
		"operationName": "questionData",
		"variables": map[string]interface{}{
			"titleSlug": slug,
		},
		"query": `query questionData($titleSlug: String!) { question(titleSlug: $titleSlug) { questionId questionFrontendId title titleSlug content difficulty likes dislikes exampleTestcases codeSnippets { lang langSlug code } topicTags { name slug } } }`,
	}
	jsonQuery, _ := json.Marshal(query) // shouldn't error here ever ...
	resp, err := http.Post(GraphQlUrl, "application/json", bytes.NewBuffer(jsonQuery))
	if err != nil {
		app.Log.Append(fmt.Sprintf("GetQuestion: could not download problem info: %s\n", err.Error()))
		return []byte{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		app.Log.Append(fmt.Sprintf("GetQuestion: could not read problem info response body: %s\n", err.Error()))
		return []byte{}, err
	}

	return body, nil
}

func (app *App) InterpretSolution(questionId int, typedCode string) ([]byte, error) {
	ques := app.fetchQuestion(app.Questions[questionId].QuestionTitleSlug)
	if ques == (question.Question{}) {
		return nil, fmt.Errorf("question %d not found", questionId)
	}

	requestBody := solution.InterpretSolutionRequest{
		DataInput:  ques.ExampleTestCases,
		Language:   "golang",
		QuestionId: strconv.Itoa(app.Questions[questionId].QuestionId),
		TypedCode:  typedCode,
	}
	jsonRequest, _ := json.Marshal(requestBody)
	requestUrlPath := BaseUrl + "problems/" + ques.TitleSlug + "/interpret_solution/"
	request, err := app.getRequestWithHeaders(http.MethodPost, requestUrlPath, jsonRequest)
	if err != nil {
		app.Log.Append(fmt.Sprintf("InterpretSolution: could not create request: %s\n", err.Error()))
		return []byte{}, err
	}
	questionUrl := BaseUrl + "problems/" + ques.TitleSlug
	request.Header.Set("Referer", questionUrl)

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		app.Log.Append(fmt.Sprintf("InterpretSolution: could not get user info data: %s\n", err.Error()))
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		app.Log.Append(fmt.Sprintf("InterpretSolution: could not read response body: %s\n", err.Error()))
		return []byte{}, err
	}

	return body, nil
}

func (app *App) CheckSolution(interpretId, titleSlug string) ([]byte, error) {
	if len(interpretId) == 0 {
		app.Log.Append("CheckSolution: interpretId must not be empty")
		return nil, fmt.Errorf("interpretId must not be empty")
	}

	requestPath := BaseUrl + "submissions/detail/" + interpretId + "/check/"
	request, err := app.getRequestWithHeaders(http.MethodGet, requestPath, nil)
	if err != nil {
		app.Log.Append(fmt.Sprintf("CheckSolution: could not create request: %s\n", err.Error()))
		return []byte{}, err
	}
	questionUrl := BaseUrl + "problems/" + titleSlug
	request.Header.Set("Referer", questionUrl)

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		app.Log.Append(fmt.Sprintf("CheckSolution: could not get check solution for given interpretId: %s\n", err.Error()))
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		app.Log.Append(fmt.Sprintf("CheckSolution: could not read solution status response body: %s\n", err.Error()))
		return []byte{}, err
	}

	return body, nil
}

func (app *App) GetUser() ([]byte, error) {
	if len(app.UserAuth.AuthCookies) == 0 {
		app.Log.Append("Authenticate first!")
		return []byte{}, fmt.Errorf("user needs to be authenticated")
	}

	query := map[string]interface{}{
		"operationName": "globalData",
		"variables":     map[string]interface{}{},
		"query":         `query globalData {  userStatus { realName username } }`,
	}
	jsonQuery, _ := json.Marshal(query)
	request, err := app.getRequestWithHeaders(http.MethodPost, GraphQlUrl, jsonQuery)
	if err != nil {
		app.Log.Append(fmt.Sprintf("GetUser: could not create request: %s\n", err.Error()))
		return []byte{}, err
	}

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		app.Log.Append(fmt.Sprintf("GetUser: could not get user info data: %s\n", err.Error()))
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		app.Log.Append(fmt.Sprintf("GetUser: could not read user metadata response body: %s\n", err.Error()))
		return []byte{}, err
	}

	return body, nil
}

func (app *App) getRequestWithHeaders(method, route string, req []byte) (*http.Request, error) {
	var body io.Reader
	if req == nil {
		body = nil
	} else {
		body = bytes.NewBuffer(req)
	}

	request, err := http.NewRequest(method, route, body)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Cookie", app.UserAuth.AuthCookies)
	request.Header.Set("X-Csrftoken", app.UserAuth.CsrfToken)
	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	request.Header.Set("Access-Control-Allow-Origin", "*")

	return request, nil
}
