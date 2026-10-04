package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/thomhuang/local-leetcode/internal/assistant"
	"github.com/thomhuang/local-leetcode/internal/question"
	"github.com/thomhuang/local-leetcode/util"
)

// RunREPL starts the line-based command prompt.
func (app *App) RunREPL() error {
	app.log().Info("repl.start", "starting command prompt")
	err := app.repl(os.Stdin, os.Stdout)
	if err != nil {
		app.log().Error("repl.stop", "command prompt exited with an error", "error", err.Error())
	} else {
		app.log().Info("repl.stop", "command prompt closed")
	}
	return err
}

func (app *App) repl(in io.Reader, out io.Writer) error {
	reader := bufio.NewReader(in)
	fmt.Fprint(out, replBanner())

	for {
		fmt.Fprint(out, accentStyle.Render("> "))
		line, err := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		if line != "" {
			if quit := app.handleREPLCommand(line, reader, out); quit {
				return nil
			}
		}
		if err != nil {
			if err != io.EOF {
				return err
			}
			fmt.Fprintln(out)
			return nil
		}
	}
}

func replBanner() string {
	return "local-leetcode\nType 'help' for commands, 'quit' to exit.\n\n"
}

// handleREPLCommand parses and runs one command line. It returns true when the
// user asked to quit.
func (app *App) handleREPLCommand(line string, in *bufio.Reader, out io.Writer) (quit bool) {
	fields := strings.Fields(line)
	command := strings.ToLower(fields[0])
	args := fields[1:]

	switch command {
	case "help", "?":
		printREPLHelp(out)
	case "quit", "exit", "q":
		return true
	case "list", "ls":
		app.replList(args, out)
	case "browse", "b":
		app.replBrowse(args, out)
	case "add":
		app.replAdd(args, out)
	case "import":
		app.replImport(args, out)
	case "auth", "authenticate":
		app.replAuth(in, out)
	case "test", "t":
		app.replTest(args, out)
	case "submit", "s":
		app.replSubmit(args, out)
	case "ask", "chat":
		app.replAsk(args, out)
	case "debug":
		app.replDebug(args, out)
	default:
		fmt.Fprintf(out, "Unknown command %q. Type 'help' for the list.\n", command)
	}
	return false
}

func printREPLHelp(out io.Writer) {
	fmt.Fprint(out, `Commands:
  list [set] [category] [status]   print a checklist (no set lists the sets)
  browse <set>                     open the interactive checklist
  add <number>                     fetch a problem and write its files
  import <set>                     import every missing problem in a set
  auth                             paste LeetCode cookies to sign in
  test <number>                    run the example test cases
  submit <number>                  submit a solution
  ask <number>                     open the assistant for a problem
  debug [level|clear|verbose on|off]   show or manage the session log
  help                             show this help
  quit                             exit
`)
}

// replSetArg returns the first argument as a set name, or explains the usage and
// lists the available sets.
func (app *App) replSetArg(args []string, out io.Writer, command string) (string, bool) {
	if len(args) == 0 {
		fmt.Fprintf(out, "Usage: %s <set>\n", command)
		app.printAvailableSets(out)
		return "", false
	}
	return args[0], true
}

// replProblemArg parses the first argument as a problem number.
func replProblemArg(args []string, out io.Writer, usage string) (int, bool) {
	if len(args) == 0 {
		fmt.Fprintf(out, "Usage: %s\n", usage)
		return 0, false
	}
	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(out, "%q is not a problem number.\n", args[0])
		return 0, false
	}
	return id, true
}

func (app *App) printAvailableSets(out io.Writer) {
	names := availableProblemSets()
	if len(names) == 0 {
		fmt.Fprintf(out, "No problem sets found under %s.\n", problemSetsDir)
		return
	}
	fmt.Fprintln(out, "Available problem sets:")
	for _, name := range names {
		fmt.Fprintln(out, "  "+name)
	}
}

func (app *App) replList(args []string, out io.Writer) {
	if len(args) == 0 {
		app.printAvailableSets(out)
		return
	}
	category, status := "", ""
	if len(args) > 1 {
		category = args[1]
	}
	if len(args) > 2 {
		status = args[2]
	}
	if err := app.printProblemSet(out, args[0], category, status); err != nil {
		fmt.Fprintln(out, "list failed:", err)
	}
}

func (app *App) replBrowse(args []string, out io.Writer) {
	name, ok := app.replSetArg(args, out, "browse")
	if !ok {
		return
	}
	set, err := app.LoadProblemSet(name)
	if err != nil {
		fmt.Fprintln(out, "browse failed:", err)
		return
	}
	if len(set.Problems()) == 0 {
		fmt.Fprintf(out, "Problem set %q has no problems.\n", name)
		return
	}
	if err := runScreen(newScreenHost(app, newBrowseModel(app, set))); err != nil {
		fmt.Fprintln(out, "browse error:", err)
	}
}

func (app *App) replAdd(args []string, out io.Writer) {
	id, ok := replProblemArg(args, out, "add <number>")
	if !ok {
		return
	}
	message, err := app.AddQuestion(id)
	if err != nil {
		fmt.Fprintln(out, "add failed:", err)
		return
	}
	fmt.Fprintln(out, message)
}

func (app *App) replImport(args []string, out io.Writer) {
	name, ok := app.replSetArg(args, out, "import")
	if !ok {
		return
	}
	set, err := app.LoadProblemSet(name)
	if err != nil {
		fmt.Fprintln(out, "import failed:", err)
		return
	}

	summary := app.ImportProblemSet(set)
	fmt.Fprintf(out, "%s (%d problems)\n", set.Name, len(set.Problems()))
	fmt.Fprintf(out, "  %d added\n", summary.Added)
	fmt.Fprintf(out, "  %d skipped (already present)\n", summary.Skipped)
	fmt.Fprintf(out, "  %d premium placeholders\n", summary.Premium)
	fmt.Fprintf(out, "  %d failed\n", len(summary.Failed))
	for _, problem := range summary.Failed {
		fmt.Fprintf(out, "  failed: %s (%s)\n", problem.Title, problem.Slug)
	}
}

func (app *App) replAuth(in *bufio.Reader, out io.Writer) {
	fmt.Fprint(out, "Paste your authenticated LeetCode cookies: ")
	line, _ := in.ReadString('\n')
	session, csrfToken, ok := parseCookies(strings.TrimSpace(line))
	if !ok {
		fmt.Fprintln(out, "Cookies must contain LEETCODE_SESSION and csrftoken.")
		return
	}

	user, err := app.Authenticate(session, csrfToken)
	if err != nil {
		fmt.Fprintln(out, "Authentication failed:", err)
		return
	}
	fmt.Fprintf(out, "Signed in as %s (@%s).\n", user.Data.UserStatus.FullName, user.Data.UserStatus.Username)
}

func (app *App) replTest(args []string, out io.Writer) {
	id, ok := replProblemArg(args, out, "test <number>")
	if !ok {
		return
	}
	text, err := app.TestProblem(id)
	if err != nil {
		fmt.Fprintln(out, "test failed:", err)
		return
	}
	fmt.Fprintln(out, text)
}

func (app *App) replSubmit(args []string, out io.Writer) {
	id, ok := replProblemArg(args, out, "submit <number>")
	if !ok {
		return
	}
	text, err := app.SubmitProblem(id)
	if err != nil {
		fmt.Fprintln(out, "submit failed:", err)
		return
	}
	fmt.Fprintln(out, text)
}

func (app *App) replAsk(args []string, out io.Writer) {
	id, ok := replProblemArg(args, out, "ask <number>")
	if !ok {
		return
	}
	if !hasAssistantKey() {
		fmt.Fprintf(out, "Set %s to use the assistant.\n", deepseekAPIKeyEnv)
		return
	}
	meta, ok := app.Questions[id]
	if !ok {
		fmt.Fprintf(out, "Problem %d was not found in the problem list.\n", id)
		return
	}

	cfg, prompt, err := assistant.Load(assistantDir)
	if err != nil {
		fmt.Fprintln(out, "Unable to load assistant config:", err)
		return
	}

	problem := question.ProblemSetProblem{ID: id, Slug: meta.QuestionTitleSlug, Title: problemTitle(meta.QuestionTitleSlug)}
	system := strings.TrimRight(prompt, "\n")
	if context := buildProblemContext(problem, ""); strings.TrimSpace(context) != "" {
		system += "\n\n" + context
	}

	title := fmt.Sprintf("%d. %s", id, problem.Title)
	chat := newChatModel(app, cfg, system, title, 0, 0)
	if err := runScreen(newScreenHost(app, chatHost{chat: chat})); err != nil {
		fmt.Fprintln(out, "assistant error:", err)
	}
}

func (app *App) replDebug(args []string, out io.Writer) {
	log := app.log()

	if len(args) > 0 {
		switch strings.ToLower(args[0]) {
		case "clear":
			log.Clear()
			fmt.Fprintln(out, "Debug log cleared.")
			return
		case "verbose":
			on := len(args) > 1 && (strings.EqualFold(args[1], "on") || strings.EqualFold(args[1], "true"))
			log.SetVerbose(on)
			fmt.Fprintf(out, "Verbose logging: %v\n", on)
			return
		}
	}

	level := util.LevelDebug
	if len(args) > 0 {
		switch strings.ToLower(args[0]) {
		case "debug":
			level = util.LevelDebug
		case "info":
			level = util.LevelInfo
		case "warn":
			level = util.LevelWarn
		case "error":
			level = util.LevelError
		default:
			fmt.Fprintf(out, "Unknown debug option %q. Use: clear, verbose on|off, debug|info|warn|error.\n", args[0])
			return
		}
	}

	shown := 0
	for _, entry := range log.Entries() {
		if entry.Level < level {
			continue
		}
		fmt.Fprintln(out, debugEntryText(entry))
		shown++
	}
	fmt.Fprintf(out, "(%d entries shown)\n", shown)
}

// problemTitle humanizes a slug when no title is available, so the assistant
// context reads "Two Sum" rather than "two-sum".
func problemTitle(slug string) string {
	words := strings.Split(slug, "-")
	for i, word := range words {
		if word == "" {
			continue
		}
		words[i] = strings.ToUpper(word[:1]) + word[1:]
	}
	return strings.Join(words, " ")
}
