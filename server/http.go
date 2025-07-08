package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/thomhuang/local-leetcode/util"
	"io"
	"net/http"
)

type HttpMethod string

const (
	GET  HttpMethod = "GET"
	POST            = "POST"
)

func NewHttpServer() *HttpServer {
	return &HttpServer{
		Log: util.NewLog(),
	}
}

func (server *HttpServer) GetAllQuestions() (stream []byte, err error) {
	resp, err := http.Get("https://leetcode.com/api/problems/all/")
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
	resp, err := http.Post("https://leetcode.com/graphql", "application/json", bytes.NewBuffer(jsonQuery))
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
	request, err := server.getRequestWithHeaders(POST, "https://leetcode.com/graphql/", query)
	if err != nil {
		server.Log.Append(fmt.Sprintf("GetAllQuestions: could not create request: %s\n", err.Error()))
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

func (server *HttpServer) getRequestWithHeaders(method, route string, req map[string]interface{}) (*http.Request, error) {
	jsonQuery, _ := json.Marshal(req) // shouldn't error here ever, always will give valid json structure
	request, err := http.NewRequest(method, route, bytes.NewBuffer(jsonQuery))
	if err != nil {
		return nil, err
	}

	request.Header.Set("Cookie", server.UserAuth.AuthCookies)
	request.Header.Set("Content-Type", "application/json")

	return request, nil
}
