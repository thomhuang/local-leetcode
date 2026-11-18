package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/thomhuang/local-leetcode/internal/user"
)

func TestLoadQuestionsFromMergesCacheSetsAndLocalFiles(t *testing.T) {
	dir := t.TempDir()
	cachePath := filepath.Join(dir, "all_problems.json")
	setsDir := filepath.Join(dir, "ProblemSets")
	probsDir := filepath.Join(dir, "problems")

	cacheJSON := `{"stat_status_pairs":[{"stat":{"question_id":1,"question__title_slug":"two-sum","frontend_question_id":1}}]}`
	if err := os.WriteFile(cachePath, []byte(cacheJSON), filePerm); err != nil {
		t.Fatalf("write cache: %v", err)
	}

	if err := os.MkdirAll(setsDir, dirPerm); err != nil {
		t.Fatalf("mkdir sets: %v", err)
	}
	setJSON := `{"name":"Test","slug":"test","categories":[{"name":"Tries","problems":[{"id":208,"title":"Implement Trie","slug":"implement-trie-prefix-tree"}]}]}`
	if err := os.WriteFile(filepath.Join(setsDir, "test.json"), []byte(setJSON), filePerm); err != nil {
		t.Fatalf("write set: %v", err)
	}

	localDir := filepath.Join(probsDir, "lru-cache")
	if err := os.MkdirAll(localDir, dirPerm); err != nil {
		t.Fatalf("mkdir local problem: %v", err)
	}
	if err := os.WriteFile(filepath.Join(localDir, "146-lru-cache.go"), []byte("package lru_cache\n"), filePerm); err != nil {
		t.Fatalf("write local problem: %v", err)
	}

	app := &App{}
	got := app.loadQuestionsFrom(cachePath, setsDir, probsDir)
	for id, wantSlug := range map[int]string{
		1:   "two-sum",
		208: "implement-trie-prefix-tree",
		146: "lru-cache",
	} {
		if got[id].QuestionTitleSlug != wantSlug {
			t.Fatalf("got[%d] = %+v, want slug %q", id, got[id], wantSlug)
		}
	}
}

func TestResolveQuestionUsesGraphQLWhenUncached(t *testing.T) {
	var gotReferer string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotReferer = r.Header.Get("Referer")
		var payload struct {
			OperationName string `json:"operationName"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("decode request: %v", err)
		}
		if payload.OperationName != "problemsetQuestionList" {
			t.Errorf("operationName = %q, want problemsetQuestionList", payload.OperationName)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"problemsetQuestionList":{"total":2,"questions":[
			{"frontendQuestionId":"350","title":"Intersection","titleSlug":"intersection-of-two-arrays-ii"},
			{"frontendQuestionId":"3500","title":"New Problem","titleSlug":"new-problem"}
		]}}}`))
	}))
	defer srv.Close()

	app := &App{
		client:  srv.Client(),
		baseURL: srv.URL + "/",
	}

	meta, err := app.resolveQuestion(3500)
	if err != nil {
		t.Fatalf("resolveQuestion(3500) error = %v", err)
	}
	if meta.QuestionTitleSlug != "new-problem" {
		t.Fatalf("slug = %q, want new-problem", meta.QuestionTitleSlug)
	}
	if gotReferer == "" {
		t.Fatal("expected Referer header to be set on GraphQL request")
	}
	if app.Questions[3500].QuestionTitleSlug != "new-problem" {
		t.Fatalf("expected resolved question to be cached in app.Questions")
	}

	if _, err := app.resolveQuestion(9999); err == nil {
		t.Fatal("resolveQuestion(9999) error = nil, want not-found error")
	}
}

func TestFetchUserRejectsUnsignedResponse(t *testing.T) {
	signedIn := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if !signedIn {
			_, _ = w.Write([]byte(`{"data":{"userStatus":{"isSignedIn":false,"realName":"","username":""}}}`))
			return
		}
		_, _ = w.Write([]byte(`{"data":{"userStatus":{"isSignedIn":true,"realName":"Alice","username":"alice"}}}`))
	}))
	defer srv.Close()

	app := &App{
		UserAuth: user.UserAuthInfo{AuthCookies: "LEETCODE_SESSION=expired; csrftoken=tok", CsrfToken: "tok"},
		client:   srv.Client(),
		baseURL:  srv.URL + "/",
	}

	if _, err := app.fetchUser(); err == nil {
		t.Fatal("fetchUser() with unsigned status returned nil error, want error")
	}

	signedIn = true
	resp, err := app.fetchUser()
	if err != nil {
		t.Fatalf("fetchUser() signed-in error = %v", err)
	}
	if resp.Data.UserStatus.Username != "alice" {
		t.Fatalf("username = %q, want alice", resp.Data.UserStatus.Username)
	}
}
