package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	directory = "output/auth/"
	fileName  = "leetcode_auth.json"
	fullPath  = directory + fileName
)

func (server *HttpServer) ImportAuthentication() error {
	err := os.MkdirAll(directory, os.ModeDir)
	if err != nil {
		return fmt.Errorf("failed to create auth directory: %w", err)
	}

	authFileStream, err := os.ReadFile(fullPath)
	if err != nil {
		return fmt.Errorf("failed to read auth file: %w", err)
	}

	err = json.Unmarshal(authFileStream, &server.UserAuth)
	if err != nil {
		return fmt.Errorf("failed to parse auth file: %w", err)
	}

	lastUpdated := server.UserAuth.LastUpdated
	// We should be authenticated for 7 days, so if already authenticated within 5, ask if they still want to ...
	fiveDaysPrev := time.Now().AddDate(0, 0, -5)
	if len(server.UserAuth.AuthCookies) == 0 || fiveDaysPrev.After(lastUpdated) {
		fmt.Println("Reminder to authenticate!")
	} else {
		user := GetUser()
		fmt.Printf("\nWelcome back, %s!\n\n", user.Data.UserStatus.Username)
	}

	return nil
}

func (server *HttpServer) SaveAuthentication(cookiePairs map[string]string) error {
	if len(cookiePairs) == 0 {
		return nil
	}

	var cookies string
	for ck, cv := range cookiePairs {
		cookies += ck + "=" + cv + "; "
	}

	server.UserAuth.AuthCookies = strings.TrimRight(cookies, "; ")
	server.UserAuth.LastUpdated = time.Now()

	authJson, err := json.Marshal(server.UserAuth)
	if err != nil {
		return fmt.Errorf("failed to marshal auth data: %w", err)
	}

	err = os.MkdirAll(directory, os.ModeDir)
	if err != nil {
		return fmt.Errorf("failed to create auth directory: %w", err)
	}

	err = os.WriteFile(fullPath, authJson, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to write auth file: %w", err)
	}

	return nil
}
