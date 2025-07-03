package src

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/thomhuang/local-leetcode/util"
)

type HttpServer struct {
	Log *util.Log
}

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
	jsonQuery, err := json.Marshal(query)
	resp, err := http.Post("https://leetcode.com/graphql", "application/json", bytes.NewBuffer(jsonQuery))
	if err != nil {
		server.Log.Append(fmt.Sprintf("GetQuestion: could not download problem info: %s\n", err.Error()))
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
