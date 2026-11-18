package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"time"

	"github.com/thomhuang/local-leetcode/internal/user"
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

	if len(app.UserAuth.AuthCookies) == 0 {
		app.log().Info("auth.load", "saved session is empty")
		fmt.Println("Reminder to authenticate!")
		return nil
	}

	status, err := app.fetchUser()
	if err != nil {
		app.log().Warn("auth.load", "saved session expired or was rejected",
			"last_updated", app.UserAuth.LastUpdated.Format(time.RFC3339),
			"error", err.Error())
		app.UserAuth = user.UserAuthInfo{}
		fmt.Println("Saved session expired or was rejected. Reminder to authenticate!")
		return nil
	}
	app.log().Info("auth.load", "restored session", "username", status.Data.UserStatus.Username)
	fmt.Printf("\nWelcome back, %s!\n\n", status.Data.UserStatus.Username)

	return nil
}

func (app *App) SaveAuthentication(session, csrfToken string) error {
	return app.SaveCookieHeader("LEETCODE_SESSION="+session+"; csrftoken="+csrfToken, csrfToken)
}

func (app *App) SaveCookieHeader(cookieHeader, csrfToken string) error {
	app.log().Debug("auth.save", "writing authentication", "path", authFile)
	app.UserAuth.AuthCookies = cookieHeader
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
