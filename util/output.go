package util

import (
	"fmt"
	"github.com/thomhuang/local-leetcode/internal"
	"os"
	"strings"
)

func SaveMarkdownContent(q internal.Question) error {
	var sb strings.Builder

	sb.WriteString("# ")
	sb.WriteString(q.Difficulty)
	sb.WriteString(": ")
	sb.WriteString(q.QuestionId)
	sb.WriteString(". ")
	sb.WriteString(q.Title)
	sb.WriteByte('\n')
	sb.WriteString(q.Content)

	dir := fmt.Sprintf("output/problems/%s/", q.TitleSlug)
	err := os.MkdirAll(dir, os.ModeDir)
	if err != nil {
		return err
	}

	err = os.WriteFile(fmt.Sprintf("%s%s.md", dir, q.TitleSlug), []byte(sb.String()), os.ModePerm)
	if err != nil {
		return err
	}
	err = os.WriteFile(fmt.Sprintf("%s%s.go", dir, q.TitleSlug), []byte(q.CodeSnippet), os.ModePerm)
	if err != nil {
		return err
	}

	fmt.Println(fmt.Sprintf("Successfully created content for %s. %s\n", q.QuestionId, q.Title))
	return nil
}
