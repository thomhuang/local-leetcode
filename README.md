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
   to LeetCode and prints the submission result.
6. **Exit.**

## Problem sets

`server/output/ProblemSets/` holds named problem sets as JSON. Each set lists
its categories and the problems inside them. `neetcode-150.json` contains all
150 problems from the [NeetCode 150](https://neetcode.io/practice).

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

## Layout

| Path | Purpose |
| --- | --- |
| `server/` | The command line program. |
| `internal/` | Request and response models for the LeetCode API. |
| `util/` | The file logger. |
| `server/output/problems/` | Your solutions. Tracked in git. |
| `server/output/ProblemSets/` | Named problem sets, for example `neetcode-150.json`. Tracked in git. |
| `server/output/all_problems.json` | Cached problem list. Refreshed every 5 days. Ignored by git. |
| `server/output/local_leetcode_logs.txt` | Log for the last run. Ignored by git. |

## Develop

```
go vet ./...
go test ./...
```
