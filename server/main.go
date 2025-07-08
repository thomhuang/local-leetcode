package main

import (
	"fmt"
	"github.com/thomhuang/local-leetcode/internal"
	"github.com/thomhuang/local-leetcode/util"
	"os"
)

type HttpServer struct {
	Log       *util.Log
	UserAuth  internal.UserAuthInfo
	Questions map[int]string
}

var server *HttpServer

func main() {
	_ = os.Mkdir("output", os.ModeDir)

	server = NewHttpServer()

	server.Questions = server.getQuestions()

	err := server.ImportAuthentication()
	if err != nil {
		server.Log.Append(fmt.Sprintf("Unable to import existing authentication! %s", err.Error()))
	}
	server.Prompt()
}
