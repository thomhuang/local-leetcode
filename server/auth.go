package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"time"
)

func (app *App) ImportAuthentication() error {
	authFileStream, err := os.ReadFile(authFile)
	if errors.Is(err, fs.ErrNotExist) {
		fmt.Println("Reminder to authenticate!")
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to read auth file: %w", err)
	}

	err = json.Unmarshal(authFileStream, &app.UserAuth)
	if err != nil {
		return fmt.Errorf("failed to parse auth file: %w", err)
	}

	// LeetCode sessions last about 7 days. Remind the user before the session ends.
	staleBefore := time.Now().AddDate(0, 0, -authFreshnessDays)
	if len(app.UserAuth.AuthCookies) == 0 || staleBefore.After(app.UserAuth.LastUpdated) {
		fmt.Println("Reminder to authenticate!")
		return nil
	}

	user, err := app.fetchUser()
	if err != nil {
		fmt.Println("Saved session was rejected. Reminder to authenticate!")
		return err
	}
	fmt.Printf("\nWelcome back, %s!\n\n", user.Data.UserStatus.Username)

	return nil
}

func (app *App) SaveAuthentication(session, csrfToken string) error {
	app.UserAuth.AuthCookies = "LEETCODE_SESSION=" + session + "; csrftoken=" + csrfToken
	app.UserAuth.CsrfToken = csrfToken
	app.UserAuth.LastUpdated = time.Now()

	authJson, err := json.Marshal(app.UserAuth)
	if err != nil {
		return fmt.Errorf("failed to marshal auth data: %w", err)
	}

	err = os.MkdirAll(authDir, dirPerm)
	if err != nil {
		return fmt.Errorf("failed to create auth directory: %w", err)
	}

	err = os.WriteFile(authFile, authJson, filePerm)
	if err != nil {
		return fmt.Errorf("failed to write auth file: %w", err)
	}

	return nil
}
