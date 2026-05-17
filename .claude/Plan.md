# QAgent MVP — 3-Day Hour-by-Hour Execution Plan

> **Paradigm:** Strict procedural Go. No interfaces. No deep OOP. Raw structs + pure top-to-bottom functions.
> **Total:** 24 hours across 3 days × 8 hours/day.

---

## DAY 1 — Foundation & Context Engine

### Hour 1 — Project Scaffold & Module Init

**Goal:** Runnable skeleton with dependency resolution locked.

**Commands:**
```bash
mkdir qagent && cd qagent
go mod init github.com/yourname/qagent
go get github.com/fatih/color@latest
go get github.com/charmbracelet/bubbletea/v2@latest
go get github.com/charmbracelet/bubbles@latest
mkdir -p internal/{loader,llm,parser,runner,ui}
touch main.go internal/loader/loader.go internal/llm/llm.go \
      internal/parser/parser.go internal/runner/runner.go internal/ui/ui.go
```

**Packages:** `os`, `fmt`, `github.com/fatih/color`

**Structs to define in `main.go`:**
```go
type Config struct {
    TargetFile   string
    ModelURL     string
    ModelName    string
    MaxHeals     int
    OutputFile   string
}

type RunResult struct {
    TestFile   string
    Passed     bool
    Attempts   int
    FinalError string
}
```

**Test command:**
```bash
go build ./... && echo "BUILD OK"
```

**Edge cases:** Confirm Go 1.21+ for `os.ReadFile`. Pin `go 1.21` in `go.mod`.

---

### Hour 2 — CLI Argument Parsing

**Goal:** Parse `--file`, `--model-url`, `--model`, `--max-heals` from `os.Args` with zero external deps.

**File:** `main.go`

**Function to write:**
```go
func parseArgs(args []string) (Config, error)
```

**Logic:** Manual `for` loop over `args`, matching `--key value` pairs. Return error string if `--file` is absent or the path doesn't exist (`os.Stat`).

**Test command:**
```bash
go run main.go --file ./testdata/sample.go --model-url http://localhost:11434/api/chat
go run main.go  # should print usage and exit 1
```

**Edge cases:**
- Flag with no value (last arg is a `--key` with nothing after it)
- File path with spaces — wrap in quotes in shell; `os.Stat` handles it fine in Go
- Relative vs absolute paths — normalize with `filepath.Abs` immediately

---

### Hour 3 — File Loader & Prompt Builder

**Goal:** Read the target `.go` file and construct the full system prompt string.

**File:** `internal/loader/loader.go`

**Structs:**
```go
type FileContext struct {
    FilePath    string
    PackageName string
    RawSource   string
    FuncNames   []string
}
```

**Functions:**
```go
func LoadFile(path string) (FileContext, error)
func ExtractPackageName(src string) string
func ExtractFuncNames(src string) []string
func BuildSystemPrompt(ctx FileContext) string
func BuildFewShotExample() string   // returns a hardcoded gold-standard test example
```

**Implementation notes for `ExtractFuncNames`:** Use `strings.Contains` + line-by-line scan for lines starting with `func `. No need for AST parsing at MVP stage.

**The system prompt template** (hardcoded string in `BuildSystemPrompt`):
```
You are an expert Go test engineer. Your ONLY job is to write a complete, compilable Go test file.

Rules:
1. Output ONLY a single ```go ... ``` code block. No explanations.
2. The test file must start with: package <package_name>
3. Import only "testing" and standard library packages.
4. Every exported function in the source must have at least one test.
5. Use table-driven tests where applicable.

--- SOURCE FILE: <filename> ---
<raw_source>
--- END SOURCE FILE ---

--- EXAMPLE OF A PERFECT TEST FILE ---
<few_shot_example>
--- END EXAMPLE ---

Write the _test.go file now.
```

**Test command:**
```bash
# create testdata/sample.go with a simple Add(a,b int) int function
go test ./internal/loader/... -v -run TestLoadFile
```

**Edge cases:**
- Empty file → return error, don't send to LLM
- File > 8000 tokens — add a `len(src) > 32000` guard and warn; truncate with a comment
- Windows CRLF line endings in source — `strings.ReplaceAll(src, "\r\n", "\n")` before processing

---

### Hour 4 — LLM Client: Local Ollama REST

**Goal:** POST to Ollama-compatible `/api/chat` endpoint, stream disabled, get raw response body.

**File:** `internal/llm/llm.go`

**Structs:**
```go
type LLMRequest struct {
    Model    string          `json:"model"`
    Messages []LLMMessage    `json:"messages"`
    Stream   bool            `json:"stream"`
}

type LLMMessage struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

type LLMResponse struct {
    Message LLMMessage `json:"message"`
    Error   string     `json:"error,omitempty"`
}
```

**Functions:**
```go
func CallLLM(url string, req LLMRequest, timeoutSec int) (string, error)
```

**Implementation:** `net/http` client with explicit `Timeout`. `json.Marshal` the request. `ioutil.ReadAll` (or `io.ReadAll`) the body. `json.Unmarshal` into `LLMResponse`.

**Test command:**
```bash
# requires Ollama running locally with gemma3 pulled
curl -s http://localhost:11434/api/chat \
  -d '{"model":"gemma3","messages":[{"role":"user","content":"say hello"}],"stream":false}'
# then:
go test ./internal/llm/... -v -run TestCallLLM
```

**Edge cases:**
- Non-200 status → wrap status code in error message
- Ollama timeout (slow model) → set `timeoutSec` to 120, make it configurable
- `"error"` field populated in response body → check and surface it
- Network refused (Ollama not running) → friendly error: "Could not connect to model at <url>. Is it running?"

---

### Hour 5 — Frontier Model Toggle (OpenAI-compatible)

**Goal:** Add a second `CallLLMOpenAI` function that speaks the OpenAI `/v1/chat/completions` API, activated by a flag.

**File:** `internal/llm/llm.go` (extend)

**Structs (OpenAI wire format):**
```go
type OpenAIRequest struct {
    Model    string          `json:"model"`
    Messages []LLMMessage    `json:"messages"`
}

type OpenAIChoice struct {
    Message LLMMessage `json:"message"`
}

type OpenAIResponse struct {
    Choices []OpenAIChoice `json:"choices"`
    Error   *OpenAIError   `json:"error,omitempty"`
}

type OpenAIError struct {
    Message string `json:"message"`
}
```

**Function:**
```go
func CallLLMOpenAI(url, apiKey string, req OpenAIRequest, timeoutSec int) (string, error)
```

**Dispatch function in `main.go`:**
```go
func callModel(cfg Config, msgs []LLMMessage) (string, error) {
    if cfg.APIKey != "" {
        // frontier path
    }
    // local path
}
```

**Test command:**
```bash
OPENAI_API_KEY=sk-xxx go run main.go --file ./testdata/sample.go \
  --model-url https://api.openai.com/v1/chat/completions --model gpt-4o
```

**Edge cases:**
- API key read from `QAGENT_API_KEY` env var, never from CLI flags (avoid shell history leaks)
- Rate limit 429 → log "Rate limited. Waiting 5s..." and retry once

---

### Hour 6 — Markdown Code Block Parser

**Goal:** Extract Go source from LLM response that may contain prose + fenced code blocks.

**File:** `internal/parser/parser.go`

**Function:**
```go
func ExtractGoBlock(raw string) (string, error)
```

**Algorithm (pure string manipulation, no regex):**
```
1. Find first occurrence of "```go" (case-sensitive)
2. Slice from index + len("```go") + 1
3. Find first occurrence of "```" in the remaining slice
4. Slice up to that index → this is your code
5. strings.TrimSpace the result
6. If result is empty → return error "no go block found"
```

**Function:**
```go
func WriteTestFile(targetPath, code string) (string, error)
```
Derives test file path: `strings.TrimSuffix(targetPath, ".go") + "_test.go"`, then `os.WriteFile`.

**Test command:**
```bash
go test ./internal/parser/... -v
# test cases: clean block, block with trailing prose, no block, empty block, multiple blocks (take first)
```

**Edge cases:**
- LLM wraps in ` ```golang ` instead of ` ```go ` → also check for this prefix
- LLM outputs raw Go without fences (rare but possible) → detect by checking if raw starts with `package ` after trimming; accept as-is
- Package declaration mismatch — after extraction, verify first non-comment line starts with `package`; warn but don't block
- Nested backticks inside code (e.g., backtick string literals) — the "find first closing ```" heuristic fails. Mitigation: scan for ` ``` ` on its own line (`\n```\n` or `\n` + ` ``` ` + EOF)

---

### Hour 7 — Test Runner & Exit Code Capture

**Goal:** Run `go test -v ./...` in the target file's directory, capture stdout/stderr, return structured result.

**File:** `internal/runner/runner.go`

**Structs:**
```go
type TestResult struct {
    Passed   bool
    Stdout   string
    Stderr   string
    ExitCode int
}
```

**Function:**
```go
func RunTests(dir string) TestResult
```

**Implementation:**
```go
cmd := exec.Command("go", "test", "-v", "-count=1", "./...")
cmd.Dir = dir
var stdout, stderr bytes.Buffer
cmd.Stdout = &stdout
cmd.Stderr = &stderr
err := cmd.Run()
exitCode := 0
if exitErr, ok := err.(*exec.ExitError); ok {
    exitCode = exitErr.ExitCode()
}
return TestResult{Passed: exitCode == 0, ...}
```

**Test command:**
```bash
# create testdata/ with a passing and a failing test
go test ./internal/runner/... -v
```

**Edge cases:**
- `-count=1` flag — MANDATORY. Disables test result caching; without it, cached `ok` on a broken test file will fool you
- `go test` not on PATH in some CI environments → check with `exec.LookPath("go")` before running
- Very long stderr (LLM generated 500 line file with all bad imports) → cap stderr fed back to LLM at 2000 chars: `stderr[:min(len(stderr), 2000)]`
- Test binary compile error vs runtime failure — both produce non-zero exit but the error format differs. Compile errors go to stderr. Runtime panics go to stdout. Capture both.

---

### Hour 8 — Self-Healing Loop (Core Orchestrator)

**Goal:** Wire loader → LLM → parser → runner → heal into the main pipeline.

**File:** `main.go` — function `RunPipeline`

**Function:**
```go
func RunPipeline(cfg Config) RunResult
```

**Pseudocode:**
```
ctx = LoadFile(cfg.TargetFile)
systemPrompt = BuildSystemPrompt(ctx)
messages = [{role:"system", content:systemPrompt}, {role:"user", content:"Generate the test file."}]

for attempt = 0; attempt <= cfg.MaxHeals; attempt++ {
    rawResponse = callModel(cfg, messages)
    code, err = ExtractGoBlock(rawResponse)
    if err → log parse error, break

    testPath = WriteTestFile(cfg.TargetFile, code)
    result = RunTests(filepath.Dir(testPath))

    if result.Passed → return RunResult{Passed:true, Attempts:attempt+1}

    // heal: append assistant response + new user message with error
    messages = append(messages,
        {role:"assistant", content:rawResponse},
        {role:"user", content: buildHealPrompt(result.Stderr)},
    )
}
return RunResult{Passed:false, Attempts:cfg.MaxHeals+1, FinalError:result.Stderr}
```

**Function:**
```go
func buildHealPrompt(stderr string) string {
    return fmt.Sprintf(
        "The test file you generated failed to compile with these errors:\n\n%s\n\nFix ONLY the errors above. Output a complete corrected ```go block.",
        stderr[:min(len(stderr), 2000)],
    )
}
```

**Test command:**
```bash
go run main.go --file ./testdata/sample.go --model-url http://localhost:11434/api/chat --model gemma3
```

---

## DAY 2 — CLI Dashboard & Integration

### Hour 9 — Spinner Component

**Goal:** Animated spinner using `bubbletea` that runs in a goroutine while the LLM call blocks.

**File:** `internal/ui/spinner.go`

**Structs:**
```go
type SpinnerModel struct {
    spinner  spinner.Model
    message  string
    done     bool
}
```

**Functions:**
```go
func NewSpinner(msg string) SpinnerModel
func (m SpinnerModel) Init() tea.Cmd
func (m SpinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd)
func (m SpinnerModel) View() string
func RunSpinner(msg string, work func() error) error
```

`RunSpinner` launches the bubbletea program, runs `work` in a goroutine, sends a `tea.Quit` message when done.

**Test command:**
```bash
go run ./cmd/spinnertest/main.go  # small harness that calls RunSpinner with a 2s sleep
```

**Edge cases:**
- Bubbletea captures stdin/stdout — if not running in a real TTY (e.g., piped output), it panics. Guard: `if !isatty.IsTerminal(os.Stdout.Fd()) { fallback to plain fmt.Println }`
- The spinner must not swallow errors from `work func() error` — return them through a channel

---

### Hour 10 — Colored Status Logger

**Goal:** Rich terminal output with consistent color coding.

**File:** `internal/ui/logger.go`

**Functions:**
```go
func LogInfo(format string, args ...any)
func LogSuccess(format string, args ...any)
func LogError(format string, args ...any)
func LogWarn(format string, args ...any)
func LogStep(step int, total int, label string)
func LogDivider()
func LogCodeBlock(title, content string)
```

**Color scheme (using `fatih/color`):**
```
Info    → color.New(color.FgCyan)
Success → color.New(color.FgGreen, color.Bold)
Error   → color.New(color.FgRed, color.Bold)
Warn    → color.New(color.FgYellow)
Step    → color.New(color.FgWhite, color.Bold)
```

**`LogCodeBlock`:** Prints a dim border, then the content truncated to 30 lines, then another border. Used to display the generated test code before writing.

**Test command:**
```bash
go test ./internal/ui/... -v -run TestLogger
# visually inspect output colors
```

---

### Hour 11 — Wire UI into Pipeline

**Goal:** Replace bare `fmt.Println` calls in `RunPipeline` with spinner + logger.

**Changes to `main.go`:**

```go
// Before LLM call:
err = ui.RunSpinner(fmt.Sprintf("Calling %s...", cfg.ModelName), func() error {
    rawResponse, callErr = callModel(cfg, messages)
    return callErr
})

// After parse:
ui.LogCodeBlock("Generated Test File", code)

// After test run:
if result.Passed {
    ui.LogSuccess("✓ Tests passed on attempt %d", attempt+1)
} else {
    ui.LogError("✗ Tests failed (attempt %d/%d)", attempt+1, cfg.MaxHeals+1)
    ui.LogCodeBlock("Compiler Errors", result.Stderr)
}
```

**Test command:**
```bash
go run main.go --file ./testdata/sample.go --model-url http://localhost:11434/api/chat --model gemma3
# observe full animated pipeline
```

---

### Hour 12 — Final Summary Banner

**Goal:** Print a structured, visually distinct summary at the end of the run.

**File:** `internal/ui/summary.go`

**Function:**
```go
func PrintSummary(r RunResult, elapsed time.Duration)
```

**Output format:**
```
╔══════════════════════════════════════╗
║         Q A G E N T  R E S U L T    ║
╠══════════════════════════════════════╣
║  Status   : ✓ PASSED                ║
║  File     : math_test.go            ║
║  Attempts : 1 / 3                   ║
║  Time     : 14.2s                   ║
╚══════════════════════════════════════╝
```

Use `fmt.Sprintf` + `color.GreenString` / `color.RedString` for the status line. Draw the box with plain ASCII (`╔`, `╗`, etc.) — they work in all modern terminals.

**Test command:**
```bash
go test ./internal/ui/... -v -run TestPrintSummary
```

---

### Hours 13–14 — End-to-End Integration Testing

**Goal:** Run the full pipeline against 3 real Go files of increasing complexity.

**Test fixtures to create in `testdata/`:**

1. `math.go` — 3 pure functions (Add, Sub, Mul). Expected: pass on attempt 1.
2. `strings_util.go` — functions using `strings` and `fmt`. Expected: pass on attempt 1 or 2.
3. `http_client.go` — function making an HTTP call. Expected: tests should mock or skip; observe heal behavior.

**Commands:**
```bash
go run main.go --file ./testdata/math.go --model-url http://localhost:11434/api/chat --model gemma3
go run main.go --file ./testdata/strings_util.go --model-url http://localhost:11434/api/chat --model gemma3
go run main.go --file ./testdata/http_client.go --model-url http://localhost:11434/api/chat --model gemma3
```

**What to observe and fix:**
- Package name mismatch in generated tests → fix `BuildSystemPrompt` to be more explicit
- Missing import in generated test → confirm heal prompt surfaces the exact missing import error from stderr
- Generated test imports non-existent packages → add explicit instruction in system prompt: "Only use packages that are imported in the source file or the standard library"

---

### Hour 15 — Makefile & Dev Tooling

**Goal:** Developer UX polish.

**Create `Makefile`:**
```makefile
.PHONY: build run test clean

build:
	go build -o bin/qagent ./...

run:
	./bin/qagent --file $(FILE) --model-url $(URL) --model $(MODEL)

test:
	go test -v -count=1 ./...

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/ testdata/*_test.go
```

**Test command:**
```bash
make build
FILE=./testdata/math.go URL=http://localhost:11434/api/chat MODEL=gemma3 make run
```

---

### Hour 16 — README & Config File Support

**Goal:** Allow config from a `qagent.json` file in addition to CLI flags. CLI flags take precedence.

**File:** `internal/loader/config.go`

**Struct:**
```go
type FileConfig struct {
    ModelURL  string `json:"model_url"`
    ModelName string `json:"model_name"`
    MaxHeals  int    `json:"max_heals"`
    APIKey    string `json:"api_key"` // read from env if blank
}
```

**Function:**
```go
func LoadFileConfig(path string) (FileConfig, error)
func MergeConfig(file FileConfig, cli Config) Config
```

**Test command:**
```bash
echo '{"model_url":"http://localhost:11434/api/chat","model_name":"gemma3"}' > qagent.json
go run main.go --file ./testdata/math.go
```

---

## DAY 3 — Hardening, Edge Cases & Release

### Hour 17 — Parser Hardening

**Goal:** Make `ExtractGoBlock` bulletproof against the 5 most common LLM output failure modes.

**Failure modes & fixes:**

| Mode | Input Pattern | Fix |
|------|--------------|-----|
| Wrong fence | ` ```golang ` | Normalize to ` ```go ` before parsing |
| No fence at all | Raw Go starting with `package` | Accept if `strings.HasPrefix(trimmed, "package")` |
| Trailing explanation | Code then prose after closing fence | Already handled — stop at first closing ` ``` ` |
| Multiple blocks | LLM outputs old+new code | Take the LAST block, not the first (healed version is last) |
| BOM / zero-width chars | Unicode junk at start | `strings.TrimLeftFunc(s, func(r rune) bool { return r == '\uFEFF' || r == '\u200B' })` |

**Test command:**
```bash
go test ./internal/parser/... -v -run TestExtractGoBlock
# add table-driven tests covering all 5 modes
```

---

### Hour 18 — Runner Hardening

**Goal:** Handle subprocess edge cases that cause silent failures.

**Additions to `runner.go`:**

```go
// 1. Timeout guard — kill go test after 60s
ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
defer cancel()
cmd := exec.CommandContext(ctx, "go", "test", "-v", "-count=1", "-timeout=30s", "./...")

// 2. Detect context deadline exceeded
if ctx.Err() == context.DeadlineExceeded {
    return TestResult{Passed: false, Stderr: "go test timed out after 60s"}
}

// 3. go vet runs implicitly — surface vet errors distinctly
if strings.Contains(stderr, "# ") && strings.Contains(stderr, "go vet") {
    result.Stderr = "[vet error] " + result.Stderr
}
```

**Also add:**
```go
func CleanTestFile(path string) error  // os.Remove — called on failed final attempt
```

**Test command:**
```bash
# create an infinite loop test to verify timeout
go test ./internal/runner/... -v -run TestTimeout
```

---

### Hour 19 — Heal Prompt Engineering

**Goal:** Improve the heal prompt to maximize first-heal success rate.

**Revised `buildHealPrompt`:**
```go
func buildHealPrompt(result TestResult) string {
    errorSection := result.Stderr
    if len(errorSection) > 2000 {
        errorSection = errorSection[:2000] + "\n... (truncated)"
    }

    return fmt.Sprintf(`The Go test file you generated failed. Here are the EXACT compiler/runtime errors:

--- ERRORS ---
%s
--- END ERRORS ---

Instructions for fixing:
1. Do NOT change the package name.
2. Do NOT add any new external dependencies.
3. Fix ONLY what the errors above indicate.
4. Output a COMPLETE, corrected ```go code block — not a diff, not a partial snippet.
5. Every test function must start with TestXxx and accept *testing.T.

Generate the corrected file now.`, errorSection)
}
```

**Test:** Run against `testdata/http_client.go` and observe if the heal prompt produces a compilable fix.

---

### Hour 20 — `--dry-run` Mode & `--no-heal` Flag

**Goal:** Two power-user flags for debugging.

**`--dry-run`:** Parse args, load file, build prompt, print it to stdout, exit. Do not call LLM.

**`--no-heal`:** Set `MaxHeals = 0`. If tests fail on first attempt, exit with code 1 immediately.

**Changes to `parseArgs`:**
```go
type Config struct {
    // ... existing fields
    DryRun  bool
    NoHeal  bool
}
```

**In `RunPipeline`:**
```go
if cfg.DryRun {
    fmt.Println(BuildSystemPrompt(ctx))
    os.Exit(0)
}
```

**Test command:**
```bash
go run main.go --file ./testdata/math.go --dry-run
go run main.go --file ./testdata/math.go --no-heal --model-url http://...
```

---

### Hour 21 — Output & Logging Options

**Goal:** `--output-dir` flag + `--quiet` flag for CI use.

**`--output-dir path`:** Write `_test.go` to a specified directory instead of alongside the source file.

**`--quiet`:** Suppress spinner and colored output; print only final JSON result to stdout for machine consumption.

**JSON output format:**
```json
{"passed": true, "attempts": 1, "test_file": "./math_test.go", "elapsed_ms": 14200}
```

**Functions:**
```go
func PrintJSON(r RunResult, elapsed time.Duration)
```

**Test command:**
```bash
go run main.go --file ./testdata/math.go --quiet | jq .
```

---

### Hour 22 — Cross-Platform & PATH Checks

**Goal:** Ensure the tool works on macOS, Linux, and Windows (WSL).

**Additions to startup in `main.go`:**
```go
func checkPrerequisites() error {
    if _, err := exec.LookPath("go"); err != nil {
        return fmt.Errorf("'go' not found on PATH. Install Go: https://go.dev/dl/")
    }
    return nil
}
```

**Windows-specific:**
- Use `filepath.ToSlash` when constructing paths for display
- `go test ./...` works on Windows; no changes needed
- `color` package auto-disables on non-ANSI terminals (CMD.EXE) — no action needed; fatih/color handles it

**Test command:**
```bash
# on Linux/macOS:
go test ./... -v
# simulate missing go binary:
PATH="" go run main.go --file ./testdata/math.go  # should print prerequisite error
```

---

### Hour 23 — Final Polish Pass

**Goal:** Audit every user-facing string. Fix all `TODO` comments. Run linter.

**Checklist:**
- [ ] All error messages start with lowercase (Go convention)
- [ ] `RunResult.FinalError` is never empty string on failure (always has a message)
- [ ] No `panic()` calls in production paths — all replaced with `return err`
- [ ] `defer cancel()` present on every `context.WithTimeout`
- [ ] `go.sum` committed
- [ ] `.gitignore` includes `bin/`, `qagent.json` (may contain API key path hints), `testdata/*_test.go`

**Commands:**
```bash
go vet ./...
go run main.go --file ./testdata/math.go       # full visual check
go run main.go --file ./testdata/math.go --quiet | jq .
```

---

### Hour 24 — Release Build & Smoke Test

**Goal:** Produce a static binary and run the final acceptance test.

**Commands:**
```bash
# Static Linux binary
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/qagent-linux-amd64 .

# macOS ARM
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o bin/qagent-darwin-arm64 .

# Run acceptance test against a real non-trivial file
go run main.go \
  --file ./testdata/strings_util.go \
  --model-url http://localhost:11434/api/chat \
  --model gemma3 \
  --max-heals 2
```

**Final acceptance criteria:**
- [ ] Binary under 15MB (stripped)
- [ ] Cold start to first LLM call under 100ms
- [ ] Spinner animates without flicker on macOS Terminal and iTerm2
- [ ] `--quiet` mode produces valid JSON parseable by `jq`
- [ ] Generated `_test.go` compiles and passes for `testdata/math.go` on first attempt

---

## Appendix: Critical Edge Cases Summary

| Domain | Edge Case | Mitigation |
|--------|-----------|------------|
| Parser | ` ```golang ` fence variant | Normalize before search |
| Parser | Multiple code blocks | Take last block on heal attempts |
| Parser | BOM / zero-width chars | `TrimLeftFunc` on unicode junk |
| Parser | No fence, raw Go | Detect `package` prefix and accept |
| Runner | Test cache returning stale `ok` | Always pass `-count=1` |
| Runner | Infinite loop in generated test | `context.WithTimeout` + `-timeout=30s` flag to `go test` |
| Runner | stderr > 2000 chars | Hard truncate before LLM re-injection |
| LLM | Local model timeout | 120s client timeout, configurable |
| LLM | API key in shell history | Read from env var only, never CLI flag |
| LLM | Rate limit 429 | Single retry after 5s sleep |
| File | CRLF line endings | Normalize to LF on load |
| File | Source > 32KB | Warn + truncate with comment |
| CLI | `--file` flag missing | Print usage, exit code 1 |
| CLI | TTY detection for spinner | `isatty` check; fall back to plain print |
