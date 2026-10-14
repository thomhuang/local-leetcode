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
	for _, pair := range strings.Split(raw, ";") {
		name, value, found := strings.Cut(strings.TrimSpace(pair), "=")
		if !found {
			continue
		}
		switch name {
		case "LEETCODE_SESSION":
			session = value
		case "csrftoken":
			csrfToken = value
		}
	}
	return session, csrfToken, session != "" && csrfToken != ""
}
