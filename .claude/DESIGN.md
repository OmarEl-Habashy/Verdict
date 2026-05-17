# QAgent — CLI Design System
# Theme: Lapis Lazuli (Steven Universe)

## Inspiration
Lapis Lazuli's palette: deep ocean blues, bright cyan highlights, pale aqua wings,
dark navy shadows, and stark white accents. The terminal should feel like deep water
with light refracting through it — calm, precise, powerful.

---

## Color Palette

| Role          | Color Name        | Hex       | fatih/color constant         | Usage                          |
|---------------|-------------------|-----------|------------------------------|--------------------------------|
| Primary       | Lapis Blue        | #1B6FD4   | FgBlue + Bold                | Step headers, borders          |
| Highlight     | Cyan Flash        | #00BFFF   | FgCyan + Bold                | Active spinner, current step   |
| Wing Aqua     | Aqua Mist         | #7ECECA   | FgHiCyan                     | Info messages, file names      |
| Deep Navy     | Ocean Floor       | #0A1628   | — (background, not set)      | Implied background             |
| Success       | Water Light       | #00E5CC   | FgHiGreen                    | Pass / success states          |
| Failure       | Cracked Gem       | #FF4560   | FgHiRed                      | Fail / error states            |
| Warning       | Storm Yellow      | #FFD166   | FgYellow                     | Heal attempts, warnings        |
| Muted         | Deep Current      | #4A7FA5   | FgHiBlack                    | Borders, dividers, dim text    |
| White         | Gem Flash         | #F0F8FF   | FgHiWhite + Bold             | Emphasis, values, filenames    |

---

## Color Objects (Go — internal/ui/theme.go)

```go
package ui

import "github.com/fatih/color"

var (
    // Primary — step headers, box borders
    ColorPrimary  = color.New(color.FgBlue, color.Bold)

    // Highlight — spinner, active step indicator
    ColorActive   = color.New(color.FgCyan, color.Bold)

    // Info — file names, package names, func lists
    ColorInfo     = color.New(color.FgHiCyan)

    // Success — passed tests, green checkmark
    ColorSuccess  = color.New(color.FgHiGreen, color.Bold)

    // Failure — failed tests, red cross
    ColorError    = color.New(color.FgHiRed, color.Bold)

    // Warning — heal attempt notices
    ColorWarn     = color.New(color.FgYellow)

    // Muted — borders, dividers, secondary text
    ColorMuted    = color.New(color.FgHiBlack)

    // Emphasis — filenames in summary, key values
    ColorEmphasis = color.New(color.FgHiWhite, color.Bold)
)
```

---

## Typography & Symbols

### Step Indicators
```
[1/4] Loading source file        ← ColorPrimary for [N/N], ColorEmphasis for label
[2/4] Calling model: gemma-3     ← model name in ColorInfo
[3/4] Writing test file          
[4/4] Running go test            
```

### Status Symbols
```
✓   — success       ColorSuccess
✗   — failure       ColorError
⚠   — warning/heal  ColorWarn
◆   — info/active   ColorActive
│   — box border    ColorMuted
```

### Spinner Frames
Use the water/wave sequence to match the aquatic theme:
```go
var SpinnerFrames = []string{"◐", "◓", "◑", "◒"}
// or wave dots:
var SpinnerFrames = []string{"⠁", "⠂", "⠄", "⡀", "⢀", "⠠", "⠐", "⠈"}
```
Spinner text color: `ColorActive` (bright cyan)

---

## Component Specs

### Step Header
```
[2/4] Calling model: google/gemma-3-27b-it
```
```go
func LogStep(step, total int, label string) {
    bracket := ColorPrimary.Sprintf("[%d/%d]", step, total)
    fmt.Printf("%s %s\n", bracket, ColorEmphasis.Sprint(label))
}
```

### Info Line (indented metadata)
```
  File: math.go | Package: mathutil | Funcs: Add, Sub, Mul, Div
```
```go
// Labels in ColorMuted, values in ColorInfo
func LogMeta(pairs ...string) // alternating key, value args
```

### Divider
```
  ──────────────────────────────────────────────────
```
```go
func LogDivider() {
    ColorMuted.Println("  " + strings.Repeat("─", 50))
}
```

### Code Block (generated file preview)
```
  ┌─ Generated Test File ─────────────────────────
  │ package mathutil
  │ 
  │ import (
  │     "testing"
  │ )
  │ ...
  └───────────────────────────────────────────────
```
```go
// Top border:    ColorPrimary  "┌─ " + title + " " + repeated "─"
// Content lines: ColorMuted    "  │ " + ColorInfo content
// Bottom border: ColorPrimary  "└─" + repeated "─"
func LogCodeBlock(title, content string, maxLines int)
```
Max lines to display: **30**. If truncated, show `... (N more lines)` in `ColorMuted`.

### Heal Warning
```
  ⚠ Healing — attempt 2/3
```
```go
func LogHeal(attempt, max int) {
    ColorWarn.Printf("  ⚠ Healing — attempt %d/%d\n\n", attempt, max)
}
```

### Success Line
```
  ✓ Tests passed on attempt 1
```
```go
ColorSuccess.Printf("  ✓ Tests passed on attempt %d\n", attempt)
```

### Failure Line
```
  ✗ Tests failed (attempt 2/3)
```
```go
ColorError.Printf("  ✗ Tests failed (attempt %d/%d)\n", attempt, max)
```

---

## Summary Banner

```
  ╔══════════════════════════════════════════╗
  ║       Q A G E N T  R E S U L T          ║
  ╠══════════════════════════════════════════╣
  ║  Status    : ✓ PASSED                   ║
  ║  File      : math_test.go               ║
  ║  Attempts  : 1 / 3 max                  ║
  ║  Time      : 14.2s                      ║
  ╚══════════════════════════════════════════╝
```

**Color rules for the banner:**
- Box drawing chars (`╔ ╗ ╚ ╝ ╠ ╣ ║ ═`) → `ColorPrimary` (Lapis Blue)
- Title `Q A G E N T  R E S U L T` → `ColorActive` (Cyan Flash)
- Label keys (`Status`, `File`, `Attempts`, `Time`) → `ColorMuted`
- Label values → `ColorEmphasis`
- Status value on PASS → `ColorSuccess`
- Status value on FAIL → `ColorError`

**Banner width:** 44 chars inner content. Fixed. Do not make it dynamic.

---

## Full Visual Flow (annotated)

```
                                                    ← blank line before start
[1/4] Loading source file                           ← ColorPrimary + ColorEmphasis
  File: math.go | Package: mathutil | Funcs: Add, Sub, Mul, Div
  ──────────────────────────────────────────────────  ← ColorMuted divider

[2/4] Calling model: google/gemma-3-27b-it          ← spinner animates here
                                                    ← blank line after spinner done
  ┌─ Generated Test File ─────────────────────────  ← ColorPrimary
  │ package mathutil                                 ← ColorMuted pipe, ColorInfo code
  │ ...
  └───────────────────────────────────────────────

[3/4] Writing test file

[4/4] Running go test
  ✓ Tests passed on attempt 1                       ← ColorSuccess

  ╔══════════════════════════════════════════╗       ← ColorPrimary border
  ║       Q A G E N T  R E S U L T          ║       ← ColorActive title
  ╠══════════════════════════════════════════╣
  ║  Status    : ✓ PASSED                   ║       ← ColorSuccess status
  ║  File      : math_test.go               ║       ← ColorEmphasis value
  ║  Attempts  : 1 / 3 max                  ║
  ║  Time      : 14.2s                      ║
  ╚══════════════════════════════════════════╝
```

---

## Spacing Rules

- **1 blank line** before each `[N/4]` step header
- **No blank line** between a step header and its metadata line
- **1 blank line** after a code block's closing `└─` border
- **2 blank lines** before the summary banner
- All content lines indented **2 spaces** from left edge
- Code block content indented **4 spaces** (2 for margin + `│ `)

---

## Non-TTY / --quiet Mode

When stdout is not a TTY or `--quiet` flag is set:
- Strip all color codes (`color.NoColor = true`)
- Strip all box-drawing and unicode symbols
- Replace `✓` with `PASS`, `✗` with `FAIL`, `⚠` with `WARN`
- Output one JSON line on completion only:
```json
{"passed":true,"file":"math_test.go","attempts":1,"elapsed_ms":14200}
```

---

## Windows / PowerShell Notes

- `fatih/color` auto-enables ANSI on Windows 10+ via `golang.org/x/sys/windows`
- Box-drawing chars (`╔ ║ ═` etc.) render correctly in Windows Terminal and PowerShell 7+
- If running in legacy CMD, `color.NoColor` will be set automatically — no action needed
- Test rendering in both Windows Terminal and PowerShell 7 before release
