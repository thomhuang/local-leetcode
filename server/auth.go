package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

func (app *App) ImportAuthentication() error {
	err := os.MkdirAll(authDir, os.ModeDir)
	if err != nil {
		return fmt.Errorf("failed to create auth directory: %w", err)
	}

	authFileStream, err := os.ReadFile(authFile)
	if err != nil {
		return fmt.Errorf("failed to read auth file: %w", err)
	}

	err = json.Unmarshal(authFileStream, &app.UserAuth)
	if err != nil {
		return fmt.Errorf("failed to parse auth file: %w", err)
	}

	lastUpdated := app.UserAuth.LastUpdated
	// We should be authenticated for 7 days, so if already authenticated within 5, ask if they still want to ...
	fiveDaysPrev := time.Now().AddDate(0, 0, -authFreshnessDays)
	if len(app.UserAuth.AuthCookies) == 0 || fiveDaysPrev.After(lastUpdated) {
		fmt.Println("Reminder to authenticate!")
	} else {
		user := app.fetchUser()
		fmt.Printf("\nWelcome back, %s!\n\n", user.Data.UserStatus.Username)
	}

	return nil
}

func (app *App) SaveAuthentication(cookiePairs map[string]string) error {
	if len(cookiePairs) == 0 {
		return nil
	}

	var cookies string
	for ck, cv := range cookiePairs {
		cookies += ck + "=" + cv + "; "
	}

	app.UserAuth.AuthCookies = strings.TrimRight(cookies, "; ")
	app.UserAuth.LastUpdated = time.Now()

	authJson, err := json.Marshal(app.UserAuth)
	if err != nil {
		return fmt.Errorf("failed to marshal auth data: %w", err)
	}

	err = os.MkdirAll(authDir, os.ModeDir)
	if err != nil {
		return fmt.Errorf("failed to create auth directory: %w", err)
	}

	err = os.WriteFile(authFile, authJson, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to write auth file: %w", err)
	}

	return nil
}
