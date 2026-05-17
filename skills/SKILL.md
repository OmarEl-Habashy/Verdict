---
name: qagent-dev
description: >
  Use this skill when working on the QAgent codebase — a deterministic AI test-generation
  CLI tool written in Go. Triggers: any task touching qagent source files, adding features,
  fixing bugs, writing tests, or extending the LLM/runner/parser pipeline.
---

# QAgent Development Skill

QAgent is a **deterministic AI testing tool** written in Go. It reads a `.go` source file,
generates unit tests via an LLM, runs them with `go test`, and self-heals on compile errors
up to a configurable maximum number of attempts.

---

## Absolute Constraints (never violate)

1. **No interfaces.** Use concrete structs only. The only exception: `bubbletea` library types.
2. **No deep OOP.** No method chains, no builder patterns. Write pure top-to-bottom functions.
3. **No external dependencies beyond the approved list.** See below.
4. **No `panic()` in production paths.** All errors must be returned.
5. **No global mutable state.** Pass `Config` and context structs explicitly.
6. **All LLM API keys come from environment variables only.** Never from CLI flags or config files.
7. **`go test` must always run with `-count=1`.** Never allow cached results.
8. **Max heal attempts: 2 (default).** Configurable but never more than 5.

---

## Approved Dependencies

```
github.com/fatih/color              — terminal color output
github.com/charmbracelet/bubbletea/v2 — spinner/TUI only
github.com/charmbracelet/bubbles    — spinner component
```

Standard library only for everything else: `net/http`, `os/exec`, `encoding/json`,
`strings`, `fmt`, `bytes`, `context`, `time`, `path/filepath`.

---

## Project Layout

```
qagent/
├── main.go                    ← Config struct, parseArgs(), RunPipeline(), main()
├── internal/
│   ├── loader/
│   │   ├── loader.go          ← LoadFile, ExtractPackageName, ExtractFuncNames
│   │   ├── prompt.go          ← BuildSystemPrompt, BuildFewShotExample
│   │   └── config.go          ← LoadFileConfig, MergeConfig
│   ├── llm/
│   │   └── llm.go             ← CallLLM (Ollama), CallLLMOpenAI, callModel dispatch
│   ├── parser/
│   │   └── parser.go          ← ExtractGoBlock, WriteTestFile
│   ├── runner/
│   │   └── runner.go          ← RunTests, CleanTestFile
│   └── ui/
│       ├── logger.go           ← LogInfo/Success/Error/Warn/Step/Divider/CodeBlock
│       ├── spinner.go          ← SpinnerModel, RunSpinner
│       └── summary.go          ← PrintSummary, PrintJSON
├── testdata/
│   ├── math.go                ← Simple pure functions (fixture)
│   ├── strings_util.go        ← String manipulation (fixture)
│   └── http_client.go         ← HTTP-dependent (fixture)
├── Makefile
├── qagent.json.example
└── go.mod
```

---

## Core Data Structs

```go
// main.go
type Config struct {
    TargetFile string
    ModelURL   string
    ModelName  string
    MaxHeals   int
    OutputDir  string
    APIKey     string  // from QAGENT_API_KEY env var
    DryRun     bool
    NoHeal     bool
    Quiet      bool
}

type RunResult struct {
    TestFile   string
    Passed     bool
    Attempts   int
    FinalError string
    ElapsedMs  int64
}

// internal/loader/loader.go
type FileContext struct {
    FilePath    string
    PackageName string
    RawSource   string
    FuncNames   []string
}

// internal/llm/llm.go
type LLMMessage struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

// internal/runner/runner.go
type TestResult struct {
    Passed   bool
    Stdout   string
    Stderr   string
    ExitCode int
}
```

---

## Pipeline Flow

```
parseArgs → checkPrerequisites → LoadFile → BuildSystemPrompt
    └─ loop (attempt 0..MaxHeals):
           callModel → ExtractGoBlock → WriteTestFile → RunTests
               ├─ passed → PrintSummary → exit 0
               └─ failed → buildHealPrompt → append to messages → continue
    └─ all attempts exhausted → PrintSummary → CleanTestFile → exit 1
```

---

## Key Function Contracts

### `ExtractGoBlock(raw string) (string, error)`
- Searches for ` ```go ` and ` ```golang ` (normalize to same)
- On heal attempts (attempt > 0): take the **last** block found, not the first
- Strips BOM and zero-width unicode chars before searching
- Falls back to accepting raw input if it starts with `package ` after trimming
- Returns error if no valid block and no package prefix detected

### `RunTests(dir string) TestResult`
- Always uses: `go test -v -count=1 -timeout=30s ./...`
- Wraps in `context.WithTimeout(ctx, 60*time.Second)` for the process itself
- Caps `Stderr` at 2000 chars before storing (prevents LLM re-injection blowup)

### `buildHealPrompt(result TestResult) string`
- Instructs LLM to output a **complete** file, not a diff
- Explicitly forbids changing the package name
- Explicitly forbids adding external dependencies
- Truncates error section to 2000 chars

### `callModel(cfg Config, msgs []LLMMessage) (string, error)`
- If `cfg.APIKey != ""` → use OpenAI-compatible path (`CallLLMOpenAI`)
- Otherwise → use Ollama path (`CallLLM`)
- Both paths: 120s HTTP client timeout

---

## UI Conventions

**Color coding:**
- Cyan → informational / in-progress
- Green Bold → success / pass
- Red Bold → failure / error
- Yellow → warning / heal attempt
- White Bold → step headers

**Spinner:** Wraps every LLM call. Degrades gracefully to plain `fmt.Println` when stdout is not a TTY.

**`--quiet` mode:** Outputs only a single JSON line on completion. No color. No spinner. Machine-readable.

**Summary box:** ASCII box-drawing chars (`╔╗╚╝╠╣║═`). Never use unicode that requires font support beyond standard terminal fonts.

---

## Testing Conventions

- Every package has `*_test.go` files covering happy path + at least 2 failure modes
- Use table-driven tests (`[]struct{ name, input, want string }`)
- No mocking frameworks — use real subprocess calls with controlled fixtures in `testdata/`
- LLM calls in tests: guarded by `t.Skip("requires local model")` unless `QAGENT_INTEGRATION=1` is set

---

## Build & Release

```bash
# Dev
go run main.go --file ./testdata/math.go --model-url http://localhost:11434/api/chat --model gemma3

# Test
go test -v -count=1 ./...

# Lint
go vet ./...

# Release binary (static, stripped)
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/qagent-linux-amd64 .
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o bin/qagent-darwin-arm64 .
```

---

## Common Pitfalls

| Pitfall | Rule |
|---------|------|
| Test cache masking broken tests | ALWAYS pass `-count=1` |
| LLM outputs `golang` fence variant | Normalize before parsing |
| Package name mismatch in generated tests | Verify first non-comment line starts with `package <name>` |
| API key leaking into shell history | Env var ONLY: `os.Getenv("QAGENT_API_KEY")` |
| Spinner panicking in non-TTY | Check `isatty` before starting bubbletea |
| Generated test importing unknown packages | System prompt must say: "only use packages imported in the source or stdlib" |
| Truncated stderr confusing the heal LLM | Always annotate truncation: append `\n...(truncated)` |
