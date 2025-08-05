package user

import "time"

type UserAuthInfo struct {
	AuthCookies string
	LastUpdated time.Time
	CsrfToken   string
}

type UserStatusResponse struct {
	Data UserStatusModel `json:"data"`
}

type UserStatusModel struct {
	UserStatus UserModel `json:"userStatus"`
}

type UserModel struct {
	FullName string `json:"realName"`
	Username string `json:"username"`
}
