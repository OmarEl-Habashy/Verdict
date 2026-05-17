# QAgent — Claude Code Instructions

## Supporting Files — Load These First

| File | Purpose |
|------|---------|
| `.claude/architecture.rules` | Full rule set with code examples — authoritative source |
| `.claude/Plan.md` | Hour-by-hour 3-day execution plan |
| `skills/SKILL.md` | Quick-reference skill card for QAgent development |

---

QAgent is a **deterministic AI test-generation CLI tool** written in Go.
It reads a `.go` source file, generates unit tests via an LLM, runs them with
`go test`, and self-heals on compile errors up to a configurable maximum attempts.

> **Read this entire file before touching any code.** Every rule here is a hard
> constraint, not a suggestion. If a rule conflicts with what seems "cleaner" or
> "more idiomatic Go", the rule wins.

---

## What This Project Is

- A focused CLI binary. Not a library. Not a framework.
- Paradigm: **strict procedural, C-style Go**. Think competitive programming.
- The codebase should read top-to-bottom with zero surprise. A junior dev should
  understand any function in under 30 seconds.

## What This Project Is Not

- Not an OOP project. No class hierarchies, no builder patterns, no method chains.
- Not a "clean architecture" project. No ports & adapters, no dependency injection containers.
- Not a generics showcase. No type parameters anywhere.

---

## Hard Rules — Violating These Requires Rewriting the Code

### 1. No interfaces in production code
```go
// ❌ NEVER
type LLMClient interface { Call(req Request) (string, error) }

// ✅ ALWAYS
type OllamaClient struct { URL string; Timeout time.Duration }
func CallOllama(c OllamaClient, req LLMRequest) (string, error) { ... }
```
**Only exception:** `bubbletea` requires `tea.Model` — permitted in `internal/ui/spinner.go` only.

### 2. No methods on domain structs
Domain structs (`Config`, `FileContext`, `TestResult`, `RunResult`) are pure data bags.
```go
// ❌ NEVER
func (r RunResult) IsSuccess() bool { return r.Passed }

// ✅ ALWAYS
func isSuccess(r RunResult) bool { return r.Passed }
```

### 3. One function, one job
No function may call the LLM AND parse the response AND write a file. Split them.
```go
// ❌ NEVER — mixed concerns
func GenerateAndWrite(cfg Config, ctx FileContext) (string, error)

// ✅ ALWAYS — separated
func CallLLM(url string, req LLMRequest, timeout int) (string, error)
func ExtractGoBlock(raw string) (string, error)
func WriteTestFile(targetPath, code string) (string, error)
```

### 4. Return errors, never panic
```go
// ❌ NEVER
if err != nil { panic(err) }

// ✅ ALWAYS — with wrapping format: "funcName: context: %w"
return FileContext{}, fmt.Errorf("loadFile: reading %s: %w", path, err)
```

### 5. No global mutable state
No package-level `var` that holds runtime state. Pass `Config` explicitly everywhere.
```go
// ❌ NEVER
var globalConfig Config

// ✅ ALWAYS
func RunPipeline(cfg Config) RunResult { ... }
```
Permitted package-level declarations: `const` values and `fatih/color` color objects only.

### 6. No regex in the parser
`internal/parser` uses zero `regexp`. Use `strings.Index`, `strings.HasPrefix`,
`strings.TrimSpace`, `strings.TrimLeftFunc` only.
```go
// ❌ NEVER
re := regexp.MustCompile("```go\n([\\s\\S]+?)```")

// ✅ ALWAYS
start := strings.Index(raw, "```go")
rest  := raw[start+6:]
end   := strings.Index(rest, "```")
code  := strings.TrimSpace(rest[:end])
```

### 7. All subprocesses are bounded by context timeout
```go
// ✅ CANONICAL PATTERN — do not deviate
ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
defer cancel()
cmd := exec.CommandContext(ctx, "go", "test", "-v", "-count=1", "-timeout=30s", "./...")
var stdout, stderr bytes.Buffer
cmd.Stdout = &stdout
cmd.Stderr = &stderr
err := cmd.Run()
if ctx.Err() == context.DeadlineExceeded {
    return TestResult{Passed: false, Stderr: "go test timed out after 60s"}
}
```

### 8. API keys from environment only
```go
// ❌ NEVER — CLI flag, config file, or hardcoded
cfg.APIKey = flagValue

// ✅ ALWAYS
cfg.APIKey = os.Getenv("QAGENT_API_KEY")
```

### 9. go test always runs with -count=1
Never omit `-count=1`. Cached test results will silently mask broken generated tests.

### 10. Stderr injected into LLM is always capped at 2000 chars
```go
const maxStderrForLLM = 2000
if len(s) > maxStderrForLLM {
    s = s[:maxStderrForLLM] + "\n... (truncated)"
}
```

### 11. Coverage profiling collects with go test -coverprofile
When `--coverage` flag is set, `RunTests` passes `-coverprofile=coverage.out` to go test.
```go
// In RunTests:
args := []string{"test", "-v", "-count=1", "-timeout=30s"}
if collectCoverage {
    args = append(args, "-coverprofile=coverage.out")
}
```
The `extractCoverage()` function parses coverage.out and returns percentage (0-100).
Coverage is displayed in PrintSummary box only if > 0 and only when tests pass.

### 12. Low coverage (<80%) triggers healing
After `RunTests()` returns with passing tests, check coverage:
```go
if lastResult.Passed && lastResult.Coverage > 0 && lastResult.Coverage < 80.0 {
    lastResult.Passed = false
    lastResult.ErrorType = "low_coverage"
}
```
Tests that compile and pass but achieve <80% coverage are re-submitted to LLM via healing loop.
This ensures iterative improvement toward branch coverage target. Use dedicated `buildCoverageHealPrompt()` that guides the LLM to add missing branches and edge cases (zero, negative, error paths, nil inputs).

### 13. No speculative abstractions (YAGNI is absolute)
Do not create types, packages, or functions for features that don't exist yet.
If there are two LLM backends today, write two concrete functions — not a provider interface.

---

## Approved External Dependencies

```
github.com/fatih/color                 — terminal colors
github.com/charmbracelet/bubbletea/v2  — spinner TUI only
github.com/charmbracelet/bubbles       — spinner component
```

Everything else uses the standard library. Do **not** add new dependencies without
explicit instruction. If you think a dependency would help, say so and wait for approval.

---

## Package Layout & Import Rules

```
qagent/
├── main.go                     ← Config, RunResult, parseArgs(), RunPipeline(), main()
├── internal/
│   ├── loader/
│   │   ├── loader.go           ← LoadFile, ExtractPackageName, ExtractFuncNames
│   │   ├── prompt.go           ← BuildSystemPrompt, BuildFewShotExample
│   │   └── config.go           ← LoadFileConfig, MergeConfig
│   ├── llm/
│   │   └── llm.go              ← CallLLM (Ollama), CallLLMOpenAI, callModel dispatch
│   ├── parser/
│   │   └── parser.go           ← ExtractGoBlock, WriteTestFile
│   ├── runner/
│   │   ├── runner.go           ← RunTests, CleanTestFile, extractCoverage
│   │   ├── classify.go         ← ClassifyError (5-category mapping)
│   │   └── diagnose.go         ← DiagnoseTestFailure, extractFirstErrorFile
│   └── ui/
│       ├── logger.go            ← LogInfo/Success/Error/Warn/Step/CodeBlock
│       ├── spinner.go           ← SpinnerModel, RunSpinner
│       └── summary.go           ← PrintSummary, PrintJSON
├── testdata/                   ← fixture .go files (never auto-edited)
├── Makefile
└── go.mod                      ← pinned to go 1.21
```

**Import boundaries (enforced — never cross these):**

| Package | May import |
|---------|-----------|
| `main` | all internal packages + `os fmt time path/filepath encoding/json` |
| `internal/loader` | `os strings fmt path/filepath` only |
| `internal/llm` | `net/http encoding/json io fmt time context` only |
| `internal/parser` | `strings fmt os path/filepath` only |
| `internal/runner` | `os/exec bytes context time strings fmt` only |
| `internal/ui` | `fmt time os fatih/color bubbletea bubbles` only |

**Phase 2 Enhancement:** `main` now imports `encoding/json` for structured logging via `LogRun()`.

`internal/ui` must **never** import `internal/llm`, `internal/runner`, or `internal/loader`.
`internal/loader` must **never** import any other internal package.
Only `main` wires packages together.

---

## Core Structs (do not restructure without explicit instruction)

```go
// main.go
type Config struct {
    TargetFile string
    ModelURL   string
    ModelName  string
    MaxHeals   int
    OutputDir  string
    APIKey     string // from QAGENT_API_KEY only
    DryRun     bool
    NoHeal     bool
    Quiet      bool
    Coverage   bool   // if true, collect test coverage with -coverprofile
}

type RunResult struct {
    TestFile   string
    Passed     bool
    Attempts   int
    FinalError string
    ElapsedMs  int64
}

type RunRecord struct {
    File      string `json:"file"`
    Model     string `json:"model"`
    Attempts  int    `json:"attempts"`
    Passed    bool   `json:"passed"`
    ErrorType string `json:"error_type,omitempty"`
    Ms        int64  `json:"ms"`
    Timestamp string `json:"timestamp"`
}
```

**Phase 2 Enhancement:** `RunRecord` captures structured telemetry to `~/.qagent/runs.jsonl` for failure analysis and pattern detection.

---

## Pipeline (read before modifying RunPipeline)

```
parseArgs → checkPrerequisites → LoadFile → BuildSystemPrompt
  └─ for attempt := 0; attempt <= cfg.MaxHeals; attempt++:
         callModel → ExtractGoBlock → WriteTestFile → RunTests
             ├─ passed  → PrintSummary → LogRun (telemetry) → os.Exit(0)
             └─ failed  → ClassifyError (smart router) → buildHealPrompt (context-aware) → append to messages → continue
  └─ exhausted → PrintSummary → CleanTestFile → LogRun (telemetry) → os.Exit(1)
```

The heal loop counter always increments. No early-break that skips the counter.
`MaxHeals` default is 2. Hard ceiling is 5 (`healCeiling` const).

**Phase 2 Enhancements:**
- `ClassifyError()` routes compilation errors into 6 categories: `missing_import`, `syntax_error`, `compilation`, `runtime_error`, `low_coverage`, `unknown`
- `buildHealPrompt()` now examines error type and injects error-specific guidance, including `buildCoverageHealPrompt()` for low-coverage cases
- `LogRun()` appends structured `RunRecord` to `~/.qagent/runs.jsonl` for analysis

**Phase 3 Enhancements (Coverage-Driven Healing):**
- **Low coverage (<80%) now triggers automatic healing**: After `RunTests()` completes with `Passed=true` and `Coverage < 80.0`, flip `Passed=false` and set `ErrorType="low_coverage"` to re-submit to LLM
- **Coverage requirements in system prompt**: Added explicit rules to force LLM to test all branches, all error paths, and edge cases — mechanically produces 80%+ coverage without healing
- **Dedicated coverage healing prompt**: `buildCoverageHealPrompt()` provides context-aware guidance focused on adding missing branches and edge cases, not fixing syntax
- **Result**: Tight feedback loop that iteratively improves test coverage until it reaches 80%+ target

---

## Testing Conventions

- Table-driven tests everywhere: `[]struct{ name, input, want string }{...}`
- No mocking frameworks. Use real fixtures in `testdata/`.
- LLM-dependent tests must be guarded:
  ```go
  if os.Getenv("QAGENT_INTEGRATION") != "1" {
      t.Skip("skipping: requires local model (set QAGENT_INTEGRATION=1)")
  }
  ```
- Every package needs tests for: happy path + at least 2 distinct failure modes.
- Run tests with: `go test -v -count=1 ./...`

---

## UI Conventions

| Color | Meaning |
|-------|---------|
| Cyan | info / in-progress |
| Green Bold | pass / success |
| Red Bold | fail / error |
| Yellow | warning / heal attempt |
| White Bold | step header |

- Spinner wraps every LLM call. Check `isTTY()` before starting bubbletea. Fall back
  to `fmt.Println` when not a TTY.
- `--quiet` flag: output one JSON line only. No spinner, no color, no banner.
- Summary box uses ASCII box-drawing chars only: `╔ ╗ ╚ ╝ ╠ ╣ ║ ═`

---

## Common Mistakes to Avoid

| Mistake | Correct approach |
|---------|-----------------|
| Omitting `-count=1` from go test | Always include it — test cache causes false passes |
| Parsing ` ```golang ` as different from ` ```go ` | Normalize both to the same before searching |
| Taking the first code block on a heal attempt | Take the **last** block — the healed version is last |
| Passing full stderr to LLM | Cap at 2000 chars, append `\n...(truncated)` |
| Reading API key from CLI flag | `os.Getenv("QAGENT_API_KEY")` only |
| Starting bubbletea in a pipe/CI context | Check `isTTY()` first |
| Adding `package` line to generated test that differs from source | Verify with `strings.HasPrefix` after extraction || Not classifying errors before healing | Use `ClassifyError()` to provide context-aware guidance |
| Swallowing stderr on final failure | Log to `~/.qagent/runs.jsonl` via `LogRun()` for pattern analysis |

**Phase 2 Patterns:**
- Smart Router pattern: Always classify errors before crafting heal prompts
- Telemetry pattern: Append immutable `RunRecord` to JSONL log; never overwrite
---

## Build Commands

```bash
# Run
go run main.go --file ./testdata/math.go --model-url http://localhost:11434/api/chat --model gemma3

# Test
go test -v -count=1 ./...

# Vet
go vet ./...

# Release
GOOS=linux  GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/qagent-linux-amd64 .
GOOS=darwin GOARCH=arm64              go build -ldflags="-s -w" -o bin/qagent-darwin-arm64 .
```

---

## Go Version

Minimum: **Go 1.21**. `go.mod` must contain `go 1.21`.
No generics. No `log/slog`. Use `internal/ui` for all user-facing output.
