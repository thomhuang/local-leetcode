package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	q "github.com/thomhuang/local-leetcode/internal/question"
	s "github.com/thomhuang/local-leetcode/internal/solution"
	"github.com/thomhuang/local-leetcode/util"
)

const (
	GET  string = "GET"
	POST        = "POST"
)

const (
	BaseUrl    = "https://leetcode.com/"
	GraphQlUrl = BaseUrl + "graphql/"
)

func NewHttpServer() *HttpServer {
	return &HttpServer{
		Log: util.NewLog(),
	}
}

func (server *HttpServer) GetAllQuestions() (stream []byte, err error) {
	resp, err := http.Get(BaseUrl + "api/problems/all/")
	if err != nil {
		server.Log.Append(fmt.Sprintf("GetAllQuestions: could not download problems metadata: %s\n", err.Error()))
		return []byte{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		server.Log.Append(fmt.Sprintf("GetAllQuestions: could not read zipped problems metadata response body: %s\n", err.Error()))
		return []byte{}, err
	}

	return body, nil
}

func (server *HttpServer) GetQuestion(slug string) (stream []byte, err error) {
	if len(slug) <= 0 {
		server.Log.Append("GetQuestion: title slug must not be empty")
		return []byte{}, fmt.Errorf("slug must not be empty")
	}

	query := map[string]interface{}{
		"operationName": "questionData",
		"variables": map[string]interface{}{
			"titleSlug": slug,
		},
		"query": `query questionData($titleSlug: String!) { question(titleSlug: $titleSlug) { questionId title titleSlug content difficulty likes dislikes exampleTestcases codeSnippets { lang langSlug code } topicTags { name slug } } }`,
	}
	jsonQuery, _ := json.Marshal(query) // shouldn't error here ever ...
	resp, err := http.Post(GraphQlUrl, "application/json", bytes.NewBuffer(jsonQuery))
	if err != nil {
		server.Log.Append(fmt.Sprintf("GetQuestion: could not download problem info: %s\n", err.Error()))
		return []byte{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		server.Log.Append(fmt.Sprintf("GetQuestion: could not read zipped problems metadata response body: %s\n", err.Error()))
		return []byte{}, err
	}

	return body, nil
}

func (server *HttpServer) InterpretSolution(questionId int, typedCode string) ([]byte, error) {
	question := getQuestion(server.Questions[questionId])
	isEmpty := question == q.Question{}
	if isEmpty {
		return nil, fmt.Errorf("question %d not found", questionId)
	}

	requestBody := s.InterpretSolutionRequest{
		DataInput:  question.ExampleTestCases,
		Language:   "golang",
		QuestionId: strconv.Itoa(questionId),
		TypedCode:  typedCode,
	}
	jsonRequest, _ := json.Marshal(requestBody)
	requestUrlPath := BaseUrl + "problems/" + question.TitleSlug + "/interpret_solution/"
	request, err := server.getRequestWithHeaders(POST, requestUrlPath, jsonRequest)
	if err != nil {
		server.Log.Append(fmt.Sprintf("InterpretSolution: could not create request: %s\n", err.Error()))
		return []byte{}, err
	}
	questionUrl := BaseUrl + "problems/" + question.TitleSlug
	request.Header.Set("Referer", questionUrl)

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		server.Log.Append(fmt.Sprintf("InterpretSolution: could not get user info data: %s\n", err.Error()))
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		server.Log.Append(fmt.Sprintf("InterpretSolution: could not read response body: %s\n", err.Error()))
		return []byte{}, err
	}

	return body, nil
}

func (server *HttpServer) CheckSolution(interpretId, titleSlug string) ([]byte, error) {
	if len(interpretId) == 0 {
		errorMsg := "CheckSolution: interpretId must not valid/non-empty"
		server.Log.Append(errorMsg)
		return nil, fmt.Errorf(errorMsg)
	}

	requestPath := BaseUrl + "submissions/detail/" + interpretId + "/check/"
	request, err := server.getRequestWithHeaders(GET, requestPath, nil)
	if err != nil {
		server.Log.Append(fmt.Sprintf("CheckSolution: could not create request: %s\n", err.Error()))
		return []byte{}, err
	}
	questionUrl := BaseUrl + "problems/" + titleSlug
	request.Header.Set("Referer", questionUrl)

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		server.Log.Append(fmt.Sprintf("CheckSolution: could not get check solution for given interpretId: %s\n", err.Error()))
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		server.Log.Append(fmt.Sprintf("CheckSolution: could not read solution status response body: %s\n", err.Error()))
		return []byte{}, err
	}

	return body, nil
}

func (server *HttpServer) GetUser() (stream []byte, err error) {
	if len(server.UserAuth.AuthCookies) == 0 {
		server.Log.Append("Authenticate first!")
		return []byte{}, fmt.Errorf("user needs to be authenticated")
	}

	query := map[string]interface{}{
		"operationName": "globalData",
		"variables":     map[string]interface{}{},
		"query":         `query globalData {  userStatus { realName username } }`,
	}
	jsonQuery, _ := json.Marshal(query)
	request, err := server.getRequestWithHeaders(POST, GraphQlUrl, jsonQuery)
	if err != nil {
		server.Log.Append(fmt.Sprintf("GetUser: could not create request: %s\n", err.Error()))
		return []byte{}, err
	}

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		server.Log.Append(fmt.Sprintf("GetUser: could not get user info data: %s\n", err.Error()))
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		server.Log.Append(fmt.Sprintf("GetUser: could not read user metadata response body: %s\n", err.Error()))
		return []byte{}, err
	}

	return body, nil
}

func (server *HttpServer) getRequestWithHeaders(method, route string, req []byte) (*http.Request, error) {
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

	request.Header.Set("Cookie", server.UserAuth.AuthCookies)
	request.Header.Set("X-Csrftoken", server.UserAuth.CsrfToken)
	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	request.Header.Set("Access-Control-Allow-Origin", "*")

	return request, nil
}
