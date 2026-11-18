package main

import (
	"strings"

	"github.com/thomhuang/local-leetcode/internal/question"
)

// prepareSubmission strips the local package clause and the node types that
// the judge already supplies from a solution file so it can be sent to the
// LeetCode judge.
func prepareSubmission(source, packageName string) string {
	packageLine, _, found := strings.Cut(source, "\n")
	packageLine = strings.TrimSpace(strings.TrimPrefix(packageLine, "\uFEFF"))
	if !found || packageLine != "package "+packageName {
		return source
	}

	_, code, _ := strings.Cut(question.RemoveProvidedTypes(source), "\n")
	return strings.TrimLeft(code, "\r\n")
}

// parseCookies pulls the session and CSRF token out of a raw Cookie header.
// The last value is false when either cookie is missing.
func parseCookies(raw string) (session, csrfToken string, ok bool) {
	_, session, csrfToken, ok = normalizeCookieHeader(raw)
	return session, csrfToken, ok
}

// normalizeCookieHeader cleans a pasted Cookie header (stripping an optional
// "Cookie:" prefix and surrounding quotes) and preserves all valid cookie pairs
// alongside the extracted LEETCODE_SESSION and csrftoken values.
func normalizeCookieHeader(raw string) (cookieHeader, session, csrfToken string, ok bool) {
	s := strings.TrimSpace(strings.NewReplacer("\r", "", "\n", "").Replace(raw))
	if len(s) >= 7 && strings.EqualFold(s[:7], "cookie:") {
		s = strings.TrimSpace(s[7:])
	}
	if len(s) >= 2 && ((s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'')) {
		s = strings.TrimSpace(s[1 : len(s)-1])
	}

	var pairs []string
	for _, pair := range strings.Split(s, ";") {
		name, value, found := strings.Cut(strings.TrimSpace(pair), "=")
		name = strings.TrimSpace(name)
		value = strings.TrimSpace(value)
		if !found || name == "" || value == "" {
			continue
		}
		pairs = append(pairs, name+"="+value)
		switch name {
		case "LEETCODE_SESSION":
			session = value
		case "csrftoken":
			csrfToken = value
		}
	}
	return strings.Join(pairs, "; "), session, csrfToken, session != "" && csrfToken != ""
}

// parseSingleCookieValue extracts a bare cookie value or a single "name=value"
// assignment when the user copies individual fields from DevTools -> Cookies.
func parseSingleCookieValue(raw, wantName string) string {
	s := strings.TrimSpace(strings.Trim(strings.TrimSpace(raw), `"'`))
	s = strings.TrimSuffix(s, ";")
	if s == "" || strings.Contains(s, ";") {
		return ""
	}
	if name, value, found := strings.Cut(s, "="); found {
		if strings.EqualFold(strings.TrimSpace(name), wantName) {
			return strings.TrimSpace(value)
		}
		return ""
	}
	if strings.ContainsAny(s, " \t") {
		return ""
	}
	return s
}
