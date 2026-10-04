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
		app.log().Info("auth.load", "no saved authentication found", "path", authFile)
		fmt.Println("Reminder to authenticate!")
		return nil
	}
	if err != nil {
		app.log().Error("auth.load", "read auth file failed", "path", authFile, "error", err.Error())
		return fmt.Errorf("failed to read auth file: %w", err)
	}

	err = json.Unmarshal(authFileStream, &app.UserAuth)
	if err != nil {
		app.log().Error("auth.load", "parse auth file failed", "path", authFile, "error", err.Error())
		return fmt.Errorf("failed to parse auth file: %w", err)
	}

	// LeetCode sessions last about 7 days. Remind the user before the session ends.
	staleBefore := time.Now().AddDate(0, 0, -authFreshnessDays)
	if len(app.UserAuth.AuthCookies) == 0 || staleBefore.After(app.UserAuth.LastUpdated) {
		app.log().Info("auth.load", "saved session is stale", "last_updated", app.UserAuth.LastUpdated.Format(time.RFC3339))
		fmt.Println("Reminder to authenticate!")
		return nil
	}

	user, err := app.fetchUser()
	if err != nil {
		app.log().Error("auth.load", "saved session was rejected", "error", err.Error())
		fmt.Println("Saved session was rejected. Reminder to authenticate!")
		return err
	}
	app.log().Info("auth.load", "restored session", "username", user.Data.UserStatus.Username)
	fmt.Printf("\nWelcome back, %s!\n\n", user.Data.UserStatus.Username)

	return nil
}

func (app *App) SaveAuthentication(session, csrfToken string) error {
	app.log().Debug("auth.save", "writing authentication", "path", authFile)
	app.UserAuth.AuthCookies = "LEETCODE_SESSION=" + session + "; csrftoken=" + csrfToken
	app.UserAuth.CsrfToken = csrfToken
	app.UserAuth.LastUpdated = time.Now()

	authJson, err := json.Marshal(app.UserAuth)
	if err != nil {
		app.log().Error("auth.save", "marshal failed", "error", err.Error())
		return fmt.Errorf("failed to marshal auth data: %w", err)
	}

	err = os.MkdirAll(authDir, dirPerm)
	if err != nil {
		app.log().Error("auth.save", "create directory failed", "path", authDir, "error", err.Error())
		return fmt.Errorf("failed to create auth directory: %w", err)
	}

	err = os.WriteFile(authFile, authJson, filePerm)
	if err != nil {
		app.log().Error("auth.save", "write failed", "path", authFile, "error", err.Error())
		return fmt.Errorf("failed to write auth file: %w", err)
	}

	return nil
}
