# local-leetcode

A terminal tool that pulls LeetCode problems into local Go files and runs your
solution against the LeetCode judge.

## Run

Run the program from the repository root. All output paths are relative to it.

```
go run ./server
```

The menu offers these actions:

1. **Add a new question.** Enter a problem number. The tool writes the problem
   statement and a Go stub to `server/output/problems/<slug>/`.
2. **Add a problem set.** Choose a set from `server/output/ProblemSets/` (for
   example `neetcode-150`). The tool fetches every problem in the set that is
   not already on disk.
3. **Authenticate user.** Paste the `Cookie` header from an authenticated
   request to `https://leetcode.com/graphql`. The tool needs the
   `LEETCODE_SESSION` and `csrftoken` values. It stores them in
   `server/output/auth/`, which git ignores.
4. **Test code.** Enter a problem number. The tool sends your solution file to
   LeetCode and prints the result of the example test cases.
5. **Submit code.** Enter a problem number. The tool submits your solution file
   to LeetCode and prints the submission result. Accepted submissions are
   recorded in `server/output/progress.json`.
6. **Browse problems.** Open an interactive checklist for a problem set. See
   [Browse and progress](#browse-and-progress).
7. **Exit.**

## Problem sets

`server/output/ProblemSets/` holds named problem sets as JSON. Each set lists
its categories and the problems inside them. `neetcode-150.json` contains the
150 [NeetCode 150](https://neetcode.io/practice) problems plus an
`Additional Practice` category for other problems already in the repo.

Import a set without going through the menu:

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

The **Browse problems** action opens a full-screen checklist. Move with the
arrow keys, `enter` to open a problem, and `q`/`esc` to go back to the menu.
Inside a problem you can test, submit, read the statement, open its `.go` file
in your default editor (`o`), import it, or mark it accepted.

Each problem shows a status:

| Symbol | Meaning |
| --- | --- |
| `[x]` | Accepted on LeetCode (recorded in `progress.json`) |
| `[-]` | A solution is written but has not been accepted |
| `[ ]` | Still the starter stub |
| `[p]` | Premium; content is locked |
| `[?]` | In the set but not imported |

Use `c` to cycle the category filter and `s` to cycle the status filter while
browsing. The same checklist can be printed without the TUI, which is useful
for piping or grepping:

```
go run ./server -list neetcode-150
go run ./server -list neetcode-150 -category Graphs -status todo
```

`server/output/progress.json` stores accepted problems by slug and is tracked in
git. Submitting an accepted solution updates it automatically; you can also mark
a problem accepted by hand with `m`.

## Layout

| Path | Purpose |
| --- | --- |
| `server/` | The command line program. |
| `internal/` | Request and response models, problem sets, and progress storage. |
| `util/` | The file logger. |
| `server/output/problems/` | Your solutions. Tracked in git. |
| `server/output/ProblemSets/` | Named problem sets, for example `neetcode-150.json`. Tracked in git. |
| `server/output/progress.json` | Accepted problems by slug. Tracked in git. |
| `server/output/all_problems.json` | Cached problem list. Refreshed every 5 days. Ignored by git. |
| `server/output/local_leetcode_logs.txt` | Log for the last run. Ignored by git. |

## Develop

```
go vet ./...
go test ./...
```
