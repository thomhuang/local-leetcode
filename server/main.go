package main

import (
	"os"

	q "github.com/thomhuang/local-leetcode/internal/question"
	u "github.com/thomhuang/local-leetcode/internal/user"
	"github.com/thomhuang/local-leetcode/util"
)

type HttpServer struct {
	Log       *util.Log
	UserAuth  u.UserAuthInfo
	Questions map[int]q.QuestionMetadataModel
}

var server *HttpServer

func main() {
	_ = os.Mkdir("output", os.ModeDir)

	server = NewHttpServer()

	server.Questions = server.getQuestions()

	err := server.ImportAuthentication()
	if err != nil {
		server.Log.Append("Unable to import existing authentication! " + err.Error())
	}

	server.Prompt()
}
