# AGENTS.md

`local-leetcode` is a Go CLI and Bubble Tea TUI that pulls LeetCode problems into local Go packages, runs/submits solutions against the LeetCode judge, tracks progress, and provides a streaming DeepSeek assistant chat.

## Commands & Verification

Always run commands from the **repository root** (`server/paths.go` uses relative `"server/output/..."` paths).

- **Fast test loop (tool code only):**
  ```powershell
  go test ./server ./internal/... ./util
  ```
- **Full CI check (also type-checks all ~150 problem packages under `server/output/problems/`):**
  ```powershell
  go vet ./...
  go test ./...
  ```
- **Run CLI:** `go run ./server` (or `./scripts/run-with-bitwarden.ps1` with Bitwarden `DEEPSEEK_API_KEY`)

## Architecture & File Map

### `server/` (`package main`)
All CLI, TUI, and HTTP wiring lives in a flat `package main` centered on `App` (`server/main.go`):

- **Entry, config & REPL:**
  - `main.go`, `paths.go` — CLI flags, `App` initialization, path/timeout/permission constants.
  - `repl.go`, `prompt.go` — Line-based command prompt (`repl(io.Reader, io.Writer)` for testability), cookie parsing, and `prepareSubmission` (strips `package <slug>` and judge-supplied node types before upload).
- **Bubble Tea TUI:**
  - `screen.go`, `tui.go` — `screenHost` wrapper (attaches the global `ctrl+o` debug overlay to full-screen views; `runScreen` package var is stubbed in tests) and Lipgloss width/truncation helpers.
  - `browse.go`, `status.go`, `report.go`, `problemsets.go`, `open.go` — Interactive problem-set checklist/detail views, problem state detection (`[x]`, `[-]`, `[ ]`, `[p]`, `[?]`), plain-text `list` formatting, bulk set import, and Windows default-editor launch (`cmd /c start "" <path>`).
  - `chat.go`, `debug.go` — Streaming assistant chat view (`chatModel`) and live scrollable session-log overlay (`debugPanel`).
- **LeetCode client & persistence:**
  - `http.go`, `handlers.go`, `actions.go` — GraphQL/REST calls, generic `decode[T]`, test/submit polling, and result formatting.
  - `auth.go`, `cache.go`, `output.go` — Cookie session loading/verification, cached + on-demand GraphQL problem-list lookup, and writing problem `.md` + `.go` files.
- **DeepSeek assistant:**
  - `assistant.go`, `sse.go` — Streaming chat HTTP client (`DEEPSEEK_API_KEY` from env) and `sseReader` (`data: {...}` / `data: [DONE]`).

### `internal/` & `util/`
- `internal/question/` — LeetCode GraphQL models, HTML-to-Markdown conversion (`question_mapper.go`), problem-set JSON loader (`problem_set.go`), and AST transforms (`starter.go`: `MakeStarterCompilable` injects commented `TreeNode`/`ListNode`/`Node` structs and `panic("not implemented")`; `RemoveProvidedTypes` strips them on submit).
- `internal/solution/` — Interpret/submit/check models (`StringList` handles LeetCode fields that arrive as either string or string slice).
- `internal/progress/` — Slug-keyed accepted-problem store (`progress.json`).
- `internal/assistant/` — Assistant config/system-prompt loader and DeepSeek request/stream types.
- `internal/user/` — Auth cookie and user status models.
- `util/` — Nil-safe structured logger (`*util.Log`) backed by a 2000-entry ring buffer and `local_leetcode_logs.txt`.

## Boundaries Around `server/output/`

- **Skip `server/output/problems/` during tool searches and refactors:** It holds ~150 generated problem packages (`<slug>/<id>-<slug>.go` and `<slug>.md`). Do not modify user solutions, `server/output/progress.json`, or `server/output/ProblemSets/` unless explicitly working on problems or sets.
- **Never read, log, or commit secrets:**
  - Never read or modify `server/output/auth/` (`leetcode_auth.json` holds live session cookies).
  - Never hardcode, persist, or log `DEEPSEEK_API_KEY` or `Cookie` headers (even in verbose debug mode).
- **Gitignored runtime artifacts:** `all_problems.json`, `local_leetcode_logs.txt`, `auth/`, and `chat_logs/`.

## Coding & Testing Conventions

- **Compilable problem tree:** Generated starter files in `server/output/problems/` must remain valid Go so `go vet ./...` and `go test ./...` always pass. Any change to starter generation (`internal/question/starter.go`) must stay paired with `RemoveProvidedTypes` / `prepareSubmission`.
- **Filesystem permissions:** Always use `dirPerm` (`0o755`) and `filePerm` (`0o644`) from `server/paths.go`—never `os.ModeDir` (which creates `0000` directories on Unix CI).
- **Logging:** Log through `app.log()` (`Info`, `Debug`, `Warn`, `Error`) using `"domain.action"` event names and structured key/value pairs. Keep `*util.Log` methods nil-receiver safe so tests can instantiate `&App{}` without a logger.
- **Testable UI & I/O:** Prefer passing `io.Reader`/`io.Writer`, `httptest.Server`, `t.TempDir()`, and model `Update`/`View` calls over spawning real terminals (`runScreen`) or hitting live LeetCode/DeepSeek endpoints in tests.

## Commit Procedure

Follow **Conventional Commits** with concise, atomic commits:

```text
<type>(<scope>): <lowercase imperative summary>

<optional body paragraph explaining why the change was made>
```

- **Types:** `feat`, `fix`, `refactor`, `chore`, `docs`, `test` (omit `(<scope>)` when repo-wide, e.g., `chore: remove unreferenced helper functions`).
- **Common scopes:** `browse`, `auth`, `cache`, `assistant`, `repl`, `judge`, `output`, `progress`, `problems`, `build`.
- **Subject & body:** Use lowercase imperative mood in the subject line with no trailing period (~50–72 chars). Add a blank line and body paragraph when the motivation or tradeoff isn't obvious from the diff.
- **Staging & boundaries:**
  - Always stage explicit file paths (`git add <file>...`)—never `git add .` or `git add -A`.
  - Split independent logical changes into separate commits.
  - Never stage `server/output/problems/` or `server/output/progress.json` unless explicitly asked; when asked to commit problem work, verify those packages compile (`go vet ./... && go test ./...`) and use `feat(problems): ...` or `chore(problems): ...`.
  - Never stage `server/output/auth/` or any secrets.
- **Pre-commit check:** Run `go vet ./server ./internal/... ./util` and `go test ./server ./internal/... ./util` before committing tool changes, and do not `git push` unless explicitly asked.
