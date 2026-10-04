package main

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thomhuang/local-leetcode/util"
)

func newTestApp(t *testing.T) *App {
	t.Helper()
	log := util.NewLog(filepath.Join(t.TempDir(), "repl.log"))
	t.Cleanup(log.Close)
	return &App{Log: log}
}

func runREPL(t *testing.T, app *App, input string) string {
	t.Helper()
	var out bytes.Buffer
	if err := app.repl(strings.NewReader(input), &out); err != nil {
		t.Fatalf("repl returned error: %v", err)
	}
	return out.String()
}

func TestREPLHelpAndQuit(t *testing.T) {
	out := runREPL(t, newTestApp(t), "help\nquit\n")
	for _, want := range []string{"Commands:", "browse <set>", "ask <number>", "quit"} {
		if !strings.Contains(out, want) {
			t.Fatalf("help output missing %q:\n%s", want, out)
		}
	}
}

func TestREPLUnknownCommand(t *testing.T) {
	out := runREPL(t, newTestApp(t), "frobnicate\nquit\n")
	if !strings.Contains(out, `Unknown command "frobnicate"`) {
		t.Fatalf("missing unknown-command message:\n%s", out)
	}
}

func TestREPLProblemArgumentErrors(t *testing.T) {
	out := runREPL(t, newTestApp(t), "test\ntest abc\nquit\n")
	if !strings.Contains(out, "Usage: test <number>") {
		t.Fatalf("missing usage message:\n%s", out)
	}
	if !strings.Contains(out, `"abc" is not a problem number`) {
		t.Fatalf("missing bad-number message:\n%s", out)
	}
}

func TestREPLBrowseRequiresSet(t *testing.T) {
	out := runREPL(t, newTestApp(t), "browse\nquit\n")
	if !strings.Contains(out, "Usage: browse <set>") {
		t.Fatalf("missing browse usage:\n%s", out)
	}
}

func TestREPLAddUnknownProblem(t *testing.T) {
	out := runREPL(t, newTestApp(t), "add 42\nquit\n")
	if !strings.Contains(out, "add failed: problem 42 was not found") {
		t.Fatalf("missing add failure message:\n%s", out)
	}
}

func TestREPLDebugCommand(t *testing.T) {
	app := newTestApp(t)
	app.log().Info("test.event", "hello")

	out := runREPL(t, app, "debug\nquit\n")
	if !strings.Contains(out, "test.event") || !strings.Contains(out, "(1 entries shown)") {
		t.Fatalf("debug output missing the entry:\n%s", out)
	}

	out = runREPL(t, app, "debug error\nquit\n")
	if !strings.Contains(out, "(0 entries shown)") {
		t.Fatalf("error filter should hide the info entry:\n%s", out)
	}

	out = runREPL(t, app, "debug verbose on\nquit\n")
	if !strings.Contains(out, "Verbose logging: true") {
		t.Fatalf("verbose toggle output missing:\n%s", out)
	}

	out = runREPL(t, app, "debug clear\nquit\n")
	if !strings.Contains(out, "Debug log cleared.") {
		t.Fatalf("clear output missing:\n%s", out)
	}
}

func TestProblemTitleHumanizesSlug(t *testing.T) {
	cases := map[string]string{
		"two-sum":         "Two Sum",
		"add-two-numbers": "Add Two Numbers",
		"lru-cache":       "Lru Cache",
	}
	for slug, want := range cases {
		if got := problemTitle(slug); got != want {
			t.Fatalf("problemTitle(%q) = %q, want %q", slug, got, want)
		}
	}
}
