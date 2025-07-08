package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

const directory = "output/auth/"
const fileName = "leetcode_auth.json"
const fullPath = directory + fileName

func (server *HttpServer) ImportAuthentication() error {
	err := os.MkdirAll(directory, os.ModeDir)
	if err != nil {
		return err
	}

	authFileStream, err := os.ReadFile(fullPath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(authFileStream, &server.UserAuth)
	if err != nil {
		return err
	}

	lastUpdated := server.UserAuth.LastUpdated
	// We should be authenticated for 7 days, so if already authenticated within 5, ask if they still want to ...
	fiveDaysPrev := time.Now().AddDate(0, 0, -5)
	if len(server.UserAuth.AuthCookies) == 0 || fiveDaysPrev.After(lastUpdated) {
		fmt.Println("Reminder to authenticate!")
	} else {
		user := GetUser()
		fmt.Println(fmt.Sprintf("\nWelcome back, %s!\n", user.Data.UserStatus.Username))
	}

	return nil
}

func (server *HttpServer) SaveAuthentication(cookiePairs map[string]string) error {
	if len(cookiePairs) == 0 {
		return nil
	}

	var cookies string
	for ck, cv := range cookiePairs {
		cookies += fmt.Sprintf("%s=%s; ", ck, cv)
	}

	server.UserAuth.AuthCookies = strings.TrimRight(cookies, "; ")
	server.UserAuth.LastUpdated = time.Now()

	authJson, _ := json.Marshal(server.UserAuth)
	err := os.MkdirAll(directory, os.ModeDir)
	if err != nil {
		return err
	}

	err = os.WriteFile(fullPath, authJson, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}
