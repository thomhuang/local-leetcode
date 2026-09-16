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
2. **Authenticate user.** Paste the `Cookie` header from an authenticated
   request to `https://leetcode.com/graphql`. The tool needs the
   `LEETCODE_SESSION` and `csrftoken` values. It stores them in
   `server/output/auth/`, which git ignores.
3. **Test code.** Enter a problem number. The tool sends your solution file to
   LeetCode and prints the result of the example test cases.
4. **Submit code.** Not implemented yet.
5. **Exit.**

## Layout

| Path | Purpose |
| --- | --- |
| `server/` | The command line program. |
| `internal/` | Request and response models for the LeetCode API. |
| `util/` | The file logger. |
| `server/output/problems/` | Your solutions. Tracked in git. |
| `server/output/all_problems.json` | Cached problem list. Refreshed every 5 days. Ignored by git. |
| `server/output/local_leetcode_logs.txt` | Log for the last run. Ignored by git. |

## Develop

```
go vet ./...
go test ./...
```
