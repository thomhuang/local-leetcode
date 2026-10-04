# local-leetcode

A terminal tool that pulls LeetCode problems into local Go files and runs your
solution against the LeetCode judge.

## Run

Run the program from the repository root. All output paths are relative to it.

```
go run ./server
```

The CLI is a line-based prompt. Type a command and press `enter`; the result
prints below and the prompt returns. The full-screen views (the problem
checklist and the assistant chat) take over the terminal while they are open and
return you to the prompt when you leave them.

| Command | What it does |
| --- | --- |
| `list [set] [category] [status]` | Print a checklist. With no set, list the available sets. |
| `browse <set>` | Open the interactive checklist for a set. |
| `add <number>` | Fetch a problem and write its statement and Go stub. |
| `import <set>` | Fetch every missing problem in a set. |
| `auth` | Paste LeetCode cookies to sign in. |
| `test <number>` | Run the example test cases. |
| `submit <number>` | Submit a solution; accepted ones are recorded. |
| `ask <number>` | Open the assistant for a problem. |
| `debug [level\|clear\|verbose on\|off]` | Show or manage the session log. |
| `help` | Show the command list. |
| `quit` | Exit. |

Commands are case-insensitive, and `help` prints the same list.

`auth` needs the `Cookie` header from an authenticated request to
`https://leetcode.com/graphql`: paste the line containing `LEETCODE_SESSION` and
`csrftoken`. It stores them in `server/output/auth/`, which git ignores.

## Problem sets

`server/output/ProblemSets/` holds named problem sets as JSON. Each set lists
its categories and the problems inside them. `neetcode-150.json` contains the
150 [NeetCode 150](https://neetcode.io/practice) problems plus an
`Additional Practice` category for other problems already in the repo.

You can also import a set directly with a flag:

```
go run ./server -import-set neetcode-150
```

Importing skips problems that already exist under `server/output/problems/`, so
it is safe to run again. LeetCode Premium problems (for example
`walls-and-gates`) have no public statement or starter code, so the tool writes
a placeholder `_test.go` instead of a starter file.

Generated starter files are valid Go: they include the node types LeetCode only
shows in comments (`TreeNode`, `ListNode`, `Node`) and a `panic` body for every
function that must return a value. Replace the panic with your solution.

## Browse and progress

`browse <set>` opens the full-screen checklist for a set under
`server/output/ProblemSets/`. Move with the arrow keys, `enter` to open a
problem, and `q`/`esc` to return to the prompt.
Inside a problem you can test, submit, read the statement, open its `.go` file
in your default editor (`o`), ask the assistant (`a`), import it, or mark it
accepted.

Each problem shows a status:

| Symbol | Meaning |
| --- | --- |
| `[x]` | Accepted on LeetCode (recorded in `progress.json`) |
| `[-]` | A solution is written but has not been accepted |
| `[ ]` | Still the starter stub |
| `[p]` | Premium; content is locked |
| `[?]` | In the set but not imported |

Use `c` to cycle the category filter and `s` to cycle the status filter while
browsing. The same checklist can be printed at the prompt with `list`, which is
useful for piping or grepping:

```
go run ./server -list neetcode-150
go run ./server -list neetcode-150 -category Graphs -status todo
```

`server/output/progress.json` stores accepted problems by slug and is tracked in
git. Submitting an accepted solution updates it automatically; you can also mark
a problem accepted by hand with `m`.

## Assistant

Open the assistant with `ask <number>` from the prompt, or press `a` on a
problem's detail screen while browsing. Either way you get a chat with a
DeepSeek assistant that already knows the problem. The assistant receives the
problem statement, your current solution, and your most recent test or
submission result as context, and replies stream token by token. `enter` sends a
message, `esc` returns to the prompt, `ctrl+r` cycles the reasoning effort for
the next message (`none`/`low`/`medium`/`high`/`max`), and the arrow keys or
`pgup`/`pgdn` scroll the transcript. Press `ctrl+o` to open the debug log
without leaving the chat.

The assistant needs an API key in the `DEEPSEEK_API_KEY` environment variable:

```
DEEPSEEK_API_KEY=sk-... go run ./server
```

If you keep the key in Bitwarden, `scripts/run-with-bitwarden.ps1` fetches the
item named `DEEPSEEK API`, unlocks the vault if needed, sets `DEEPSEEK_API_KEY`
for that process only, and runs the app:

```
./scripts/run-with-bitwarden.ps1
```

That needs the Bitwarden CLI (`winget install Bitwarden.CLI`, `scoop install
bitwarden-cli`, or `npm install -g @bitwarden/cli`) and a one-time `bw login`.
The script reads the key from the item's password field, a custom field, or its
notes, and never writes it to disk.

Running with Bitwarden is optional. When the key is not set, the main menu shows
a banner and pressing `a` on a problem explains how to configure it.

The model and thinking settings live in `server/output/assistant/config.json`,
and the system prompt lives in `server/output/assistant/system_prompt.md`. Both
are created with defaults the first time you open the assistant and are tracked
in git, so you can tune the prompt without touching Go code. The default model
is `deepseek-flash`; `deepseek-v4-pro` is also available.

## Debugging

Every run keeps a structured log of what it does: HTTP requests to LeetCode,
problem imports, tests, submissions, authentication, and the assistant chat.
At the prompt, `debug` prints the session log (optionally filtered to a minimum
level, cleared, or with verbose recording toggled):

```
debug              # show every entry
debug error        # only errors
debug verbose on   # record full request/response bodies
debug clear        # empty the in-memory buffer
```

While a full-screen view is open, `ctrl+o` opens the same log as a scrollable
overlay from the browse checklist or the assistant chat.

The overlay shows the newest entries at the bottom and scrolls with the arrow
keys, `pgup`/`pgdn`, `home`, and `end`. Its controls:

| Key | Action |
| --- | --- |
| `l` | Cycle the minimum level: all, info, warn, error |
| `v` | Toggle verbose mode (records full request/response bodies) |
| `c` | Clear the in-memory buffer |
| `esc` / `ctrl+o` | Close the overlay and return to the previous screen |

Entries are colored by level: `DEBUG` (dim), `INFO` (accent), `WARN` (yellow),
and `ERROR` (red). The session buffer holds the most recent 2000 entries. The
same entries are appended to `server/output/local_leetcode_logs.txt`, which is
truncated at the start of each run and is ignored by git. Use it to inspect a
crash after the process exits; use the overlay to watch the current run live.

Assistant turns record the model, reasoning effort, message count, first-token
latency, streamed character counts, total duration, and any error. Verbose mode
additionally records the full request payload, system prompt, streamed deltas,
and final reply. The API key is never logged.

## Layout

| Path | Purpose |
| --- | --- |
| `server/` | The command line program. |
| `internal/` | Request and response models, problem sets, and progress storage. |
| `util/` | The structured logger: in-memory session buffer plus the log file. |
| `server/output/problems/` | Your solutions. Tracked in git. |
| `server/output/ProblemSets/` | Named problem sets, for example `neetcode-150.json`. Tracked in git. |
| `server/output/progress.json` | Accepted problems by slug. Tracked in git. |
| `server/output/assistant/` | Assistant config and system prompt. Tracked in git. |
| `server/output/all_problems.json` | Cached problem list. Refreshed every 5 days. Ignored by git. |
| `server/output/local_leetcode_logs.txt` | Log for the last run. Ignored by git. |

## Develop

```
go vet ./...
go test ./...
```
