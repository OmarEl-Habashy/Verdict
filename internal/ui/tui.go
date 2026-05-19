/*
Package ui manages the terminal user interfaces and logging for the application.
This file implements the interactive Text User Interface (TUI) powered by Bubble Tea,
providing a wizard to select models, directories, and specific Go files to test.

Functions:
- buildTree / walkTree: Scans directories to build a visual file tree.
- buildVisible: Determines which tree nodes are visible based on directory expansion.
- InitialModel: Sets up the initial state for the TUI wizard.
- RunTUI: Starts the Bubble Tea program and returns the user's selections.
*/
package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/OmarEl-Habashy/qagent/internal/llm"
)

var verdictArt = []string{
	`██╗   ██╗███████╗██████╗ ██████╗ ██╗ ██████╗████████╗`,
	`██║   ██║██╔════╝██╔══██╗██╔══██╗██║██╔════╝╚══██╔══╝`,
	`██║   ██║█████╗  ██████╔╝██║  ██║██║██║        ██║   `,
	`╚██╗ ██╔╝██╔══╝  ██╔══██╗██║  ██║██║██║        ██║   `,
	` ╚████╔╝ ███████╗██║  ██║██████╔╝██║╚██████╗   ██║   `,
	`  ╚═══╝  ╚══════╝╚═╝  ╚═╝╚═════╝ ╚═╝ ╚═════╝   ╚═╝  `,
}

const (
	modeProvider = "provider"
	modeModel    = "model"
	modeInput    = "input"
	modeBrowse   = "browse"
	modeConfirm  = "confirm"
)

type blinkMsg struct{}

func blinkTick() tea.Cmd {
	return tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
		return blinkMsg{}
	})
}

type inputField struct {
	value  string
	cursor int
}

func newField(s string) inputField { return inputField{value: s, cursor: len(s)} }

func (f *inputField) insert(r rune) {
	ch := string(r)
	f.value = f.value[:f.cursor] + ch + f.value[f.cursor:]
	f.cursor += len(ch)
}

func (f *inputField) deleteBack() {
	if f.cursor == 0 {
		return
	}
	_, sz := utf8.DecodeLastRuneInString(f.value[:f.cursor])
	f.value = f.value[:f.cursor-sz] + f.value[f.cursor:]
	f.cursor -= sz
}

func (f *inputField) deleteForward() {
	if f.cursor >= len(f.value) {
		return
	}
	_, sz := utf8.DecodeRuneInString(f.value[f.cursor:])
	f.value = f.value[:f.cursor] + f.value[f.cursor+sz:]
}

func (f *inputField) left() {
	if f.cursor == 0 {
		return
	}
	_, sz := utf8.DecodeLastRuneInString(f.value[:f.cursor])
	f.cursor -= sz
}

func (f *inputField) right() {
	if f.cursor >= len(f.value) {
		return
	}
	_, sz := utf8.DecodeRuneInString(f.value[f.cursor:])
	f.cursor += sz
}

func (f inputField) home() inputField { f.cursor = 0; return f }
func (f inputField) end() inputField  { f.cursor = len(f.value); return f }

func (f inputField) view(blink bool) string {
	before := f.value[:f.cursor]
	after := f.value[f.cursor:]
	if len(after) == 0 {
		if blink {
			return colorInfo.Sprint(before) + " "
		}
		return colorInfo.Sprint(before) + colorActive.Sprint("█")
	}
	_, sz := utf8.DecodeRuneInString(after)
	if blink {
		return colorInfo.Sprint(before) + colorInfo.Sprint(after[:sz]) + colorInfo.Sprint(after[sz:])
	}
	return colorInfo.Sprint(before) + colorActive.Sprint("█") + colorInfo.Sprint(after[sz:])
}

type treeNode struct {
	name     string
	fullPath string
	isDir    bool
	depth    int
	expanded bool
}

func buildTree(root string) []treeNode {
	var nodes []treeNode
	walkTree(root, 0, &nodes)
	return nodes
}

func walkTree(dir string, depth int, nodes *[]treeNode) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	var dirs, files []os.DirEntry
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".") {
			continue
		}
		if e.IsDir() {
			dirs = append(dirs, e)
		} else if strings.HasSuffix(e.Name(), ".go") && !strings.HasSuffix(e.Name(), "_test.go") {
			files = append(files, e)
		}
	}
	sort.Slice(dirs, func(i, j int) bool { return dirs[i].Name() < dirs[j].Name() })
	sort.Slice(files, func(i, j int) bool { return files[i].Name() < files[j].Name() })

	for _, d := range dirs {
		*nodes = append(*nodes, treeNode{name: d.Name(), fullPath: filepath.Join(dir, d.Name()), isDir: true, depth: depth})
		walkTree(filepath.Join(dir, d.Name()), depth+1, nodes)
	}
	for _, f := range files {
		*nodes = append(*nodes, treeNode{name: f.Name(), fullPath: filepath.Join(dir, f.Name()), depth: depth})
	}
}

func buildVisible(nodes []treeNode) []int {
	var visible []int
	collapsedDepth := -1
	for i, n := range nodes {
		if collapsedDepth >= 0 {
			if n.depth > collapsedDepth {
				continue
			}
			collapsedDepth = -1
		}
		visible = append(visible, i)
		if n.isDir && !n.expanded {
			collapsedDepth = n.depth
		}
	}
	return visible
}

type TUIModel struct {
	mode           string
	field          inputField
	rootDir        string
	tree           []treeNode
	visible        []int
	cursor         int
	selected       map[string]bool
	confirmed      bool
	err            error
	cursorBlink    bool
	providerCursor int
	provider       string
	models         []string
	modelCursor    int
	selectedModel  string
	modelInput     inputField
	inModelInput   bool
}

func InitialModel() TUIModel {
	cwd, _ := os.Getwd()
	return TUIModel{
		mode:     modeProvider,
		field:    newField(cwd),
		selected: make(map[string]bool),
	}
}

func (m TUIModel) Init() tea.Cmd { return blinkTick() }

func (m TUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if _, ok := msg.(blinkMsg); ok {
		m.cursorBlink = !m.cursorBlink
		return m, blinkTick()
	}

	k, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	m.cursorBlink = false

	switch m.mode {
	case modeProvider:
		return m.handleProvider(k)
	case modeModel:
		return m.handleModel(k)
	case modeInput:
		return m.handleInput(k)
	case modeBrowse:
		return m.handleBrowse(k)
	case modeConfirm:
		return m.handleConfirm(k)
	}
	return m, nil
}

func (m TUIModel) handleProvider(k tea.KeyMsg) (TUIModel, tea.Cmd) {
	switch k.String() {
	case "ctrl+c", "q", "esc":
		return m, tea.Quit
	case "up", "k":
		m.providerCursor = 0
	case "down", "j", "tab":
		m.providerCursor = 1
	case "enter", " ":
		if m.providerCursor == 0 {
			m.provider = "local"

			m.models = llm.GetAvailableModels("http://localhost:11434")
			m.modelCursor = 0
			if len(m.models) > 0 {
				m.selectedModel = m.models[0]
			}
			m.mode = modeModel
		} else {
			m.provider = "cloud"
			m.mode = modeInput
		}
	}
	return m, nil
}

func (m TUIModel) handleModel(k tea.KeyMsg) (TUIModel, tea.Cmd) {
	if m.inModelInput {
		return m.handleModelInput(k)
	}

	switch k.String() {
	case "ctrl+c", "q", "esc":
		return m, tea.Quit
	case "i":

		m.inModelInput = true
		m.modelInput = newField("")
		return m, nil
	case "up", "k":
		if m.modelCursor > 0 {
			m.modelCursor--
		}
	case "down", "j":
		if m.modelCursor < len(m.models)-1 {
			m.modelCursor++
		}
	case "enter", " ":
		if len(m.models) > 0 {
			m.selectedModel = m.models[m.modelCursor]
			m.mode = modeInput
		}
	case "b":
		m.mode = modeProvider
	}
	return m, nil
}

func (m TUIModel) handleModelInput(k tea.KeyMsg) (TUIModel, tea.Cmd) {
	switch k.String() {
	case "ctrl+c", "esc":
		m.inModelInput = false
		m.modelInput = inputField{}
		return m, nil
	case "enter":
		modelName := strings.TrimSpace(m.modelInput.value)
		if modelName != "" {
			m.selectedModel = modelName
			m.inModelInput = false
			m.mode = modeInput
			return m, nil
		}
		m.err = fmt.Errorf("model name cannot be empty")
		return m, nil
	case "left":
		m.modelInput.left()
	case "right":
		m.modelInput.right()
	case "home", "ctrl+a":
		m.modelInput = m.modelInput.home()
	case "end", "ctrl+e":
		m.modelInput = m.modelInput.end()
	case "backspace":
		m.modelInput.deleteBack()
	case "delete":
		m.modelInput.deleteForward()
	default:
		if k.Type == tea.KeyRunes || k.Type == tea.KeySpace {
			for _, r := range k.Runes {
				if r != 0 {
					m.modelInput.insert(r)
				}
			}
		}
	}
	return m, nil
}

func (m TUIModel) handleInput(k tea.KeyMsg) (TUIModel, tea.Cmd) {
	switch k.String() {
	case "ctrl+c", "esc":
		return m, tea.Quit
	case "enter":
		path := strings.TrimSpace(m.field.value)
		path = strings.ReplaceAll(path, "\x00", "")
		if path == "" {
			path = "."
		}
		abs, err := filepath.Abs(path)
		if err != nil {
			m.err = fmt.Errorf("invalid path: %q", path)
			return m, nil
		}
		if info, err := os.Stat(abs); err != nil || !info.IsDir() {
			m.err = fmt.Errorf("not a directory: %q (%v)", abs, err)
			return m, nil
		}
		m.err = nil
		m.rootDir = abs
		m.tree = buildTree(abs)
		m.visible = buildVisible(m.tree)
		m.cursor = 0
		m.mode = modeBrowse
	case "left":
		m.field.left()
	case "right":
		m.field.right()
	case "home", "ctrl+a":
		m.field = m.field.home()
	case "end", "ctrl+e":
		m.field = m.field.end()
	case "backspace":
		m.field.deleteBack()
		m.err = nil
	case "delete":
		m.field.deleteForward()
	default:
		if k.Type == tea.KeyRunes || k.Type == tea.KeySpace {
			for _, r := range k.Runes {
				if r != 0 {
					m.field.insert(r)
				}
			}
			m.err = nil
		}
	}
	return m, nil
}

func (m TUIModel) handleBrowse(k tea.KeyMsg) (TUIModel, tea.Cmd) {
	switch k.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "esc":
		m.mode = modeInput

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.visible)-1 {
			m.cursor++
		}

	case "right", "enter", "l":
		if len(m.visible) == 0 {
			break
		}
		idx := m.visible[m.cursor]
		n := m.tree[idx]
		if n.isDir && !n.expanded {
			m.tree[idx].expanded = true
			m.visible = buildVisible(m.tree)

			for i, vi := range m.visible {
				if vi == idx {
					m.cursor = i
					break
				}
			}
		} else if !n.isDir {

			if m.selected[n.fullPath] {
				delete(m.selected, n.fullPath)
			} else {
				m.selected[n.fullPath] = true
			}
		}

	case "left", "h":
		if len(m.visible) == 0 {
			break
		}
		idx := m.visible[m.cursor]
		n := m.tree[idx]
		if n.isDir && n.expanded {

			m.tree[idx].expanded = false
			m.visible = buildVisible(m.tree)
			for i, vi := range m.visible {
				if vi == idx {
					m.cursor = i
					break
				}
			}
		} else {

			for i := m.cursor - 1; i >= 0; i-- {
				parent := m.tree[m.visible[i]]
				if parent.isDir && parent.depth < n.depth {
					m.cursor = i
					break
				}
			}
		}

	case " ":
		if len(m.visible) == 0 {
			break
		}
		n := m.tree[m.visible[m.cursor]]
		if n.isDir {
			m.toggleDir(n.fullPath)
		} else {
			if m.selected[n.fullPath] {
				delete(m.selected, n.fullPath)
			} else {
				m.selected[n.fullPath] = true
			}
		}

	case "a":
		all := collectGoFiles(m.tree)
		if len(m.selected) == len(all) {
			m.selected = make(map[string]bool)
		} else {
			for _, f := range all {
				m.selected[f] = true
			}
		}

	case "tab":
		if len(m.selected) > 0 {
			m.mode = modeConfirm
		}
	}
	return m, nil
}

func (m TUIModel) handleConfirm(k tea.KeyMsg) (TUIModel, tea.Cmd) {
	switch k.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "enter":
		if len(m.selected) > 0 {
			m.confirmed = true
			return m, tea.Quit
		}
	case "esc", "b":
		m.mode = modeBrowse
	}
	return m, nil
}

func (m *TUIModel) toggleDir(dir string) {
	entries, _ := os.ReadDir(dir)
	var goFiles []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".go") && !strings.HasSuffix(e.Name(), "_test.go") {
			goFiles = append(goFiles, filepath.Join(dir, e.Name()))
		}
	}
	allSel := len(goFiles) > 0
	for _, f := range goFiles {
		if !m.selected[f] {
			allSel = false
			break
		}
	}
	if allSel {
		for _, f := range goFiles {
			delete(m.selected, f)
		}
	} else {
		for _, f := range goFiles {
			m.selected[f] = true
		}
	}
}

func collectGoFiles(nodes []treeNode) []string {
	var out []string
	for _, n := range nodes {
		if !n.isDir {
			out = append(out, n.fullPath)
		}
	}
	return out
}

func (m TUIModel) View() string {
	var b strings.Builder

	b.WriteString("\n")
	for _, line := range verdictArt {
		colorPrimary.Fprintf(&b, "  %s\n", line)
	}
	colorMuted.Fprintf(&b, "  %s\n", strings.Repeat("─", 55))
	colorMuted.Fprintf(&b, "  AI-powered Go test generation  ◆  v0.1\n\n")

	switch m.mode {
	case modeProvider:
		m.renderProvider(&b)
	case modeModel:
		m.renderModel(&b)
	case modeInput:
		m.renderInput(&b)
	case modeBrowse:
		m.renderBrowse(&b)
	case modeConfirm:
		m.renderConfirm(&b)
	}

	if m.err != nil {
		colorError.Fprintf(&b, "\n  ✗ %v\n", m.err)
	}
	return b.String()
}

func (m TUIModel) renderProvider(b *strings.Builder) {
	colorEmphasis.Fprintf(b, "  [1/5] Select Model Provider\n\n")

	choices := []string{"Local (Ollama / Mistral)", "Cloud (OpenRouter / Claude)"}
	for i, choice := range choices {
		if m.providerCursor == i {
			colorActive.Fprintf(b, "  ▸ %s\n", choice)
		} else {
			colorInfo.Fprintf(b, "    %s\n", choice)
		}
	}
	b.WriteString("\n")
	colorMuted.Fprintln(b, "  ↑↓ move   ↵ Enter select   q quit")
}

func (m TUIModel) renderModel(b *strings.Builder) {
	colorEmphasis.Fprintf(b, "  [2/5] Select Local Model\n\n")

	if m.inModelInput {
		m.renderModelInput(b)
		return
	}

	if len(m.models) == 0 {
		colorWarn.Fprintln(b, "  ⚠ No models found locally")
		colorMuted.Fprintln(b, "  Pull a model to Ollama first, or enter one manually\n")
		colorMuted.Fprintln(b, "  ↵ Enter continue   i type model   B back   q quit")
		return
	}

	const maxVisible = 15
	start := 0
	if m.modelCursor >= maxVisible {
		start = m.modelCursor - maxVisible + 1
	}
	end := start + maxVisible
	if end > len(m.models) {
		end = len(m.models)
	}

	for i := start; i < end; i++ {
		if i == m.modelCursor {
			colorActive.Fprintf(b, "  ▸ %s\n", m.models[i])
		} else {
			colorInfo.Fprintf(b, "    %s\n", m.models[i])
		}
	}

	if len(m.models) > maxVisible {
		colorMuted.Fprintf(b, "\n  ... %d more\n", len(m.models)-maxVisible)
	}

	b.WriteString("\n")
	colorMuted.Fprintln(b, "  ↑↓ move   ↵ Enter select   i type model   B back   q quit")
}

func (m TUIModel) renderModelInput(b *strings.Builder) {
	colorEmphasis.Fprintf(b, "  Enter model name (e.g. qwen2.5-coder:1.5b, mistral)\n\n")
	colorMuted.Fprintf(b, "  ▶ ")
	b.WriteString(m.modelInput.view(m.cursorBlink))
	b.WriteString("\n\n")

	if m.err != nil {
		colorError.Fprintf(b, "  ✗ %v\n\n", m.err)
	}

	colorMuted.Fprintln(b, "  ← →  move cursor    Home/End  jump    ↵ Enter  confirm    ESC  cancel")
}

func (m TUIModel) renderInput(b *strings.Builder) {
	colorEmphasis.Fprintf(b, "  [3/5] Enter directory path\n\n")
	colorMuted.Fprintf(b, "  ▶ ")
	b.WriteString(m.field.view(m.cursorBlink))
	b.WriteString("\n\n")
	colorMuted.Fprintln(b, "  ← →  move cursor    Home/End  jump    ↵ Enter  open    ESC  quit")
}

func (m TUIModel) renderBrowse(b *strings.Builder) {
	colorEmphasis.Fprintf(b, "  [4/5] Select files\n\n")
	colorMuted.Fprintf(b, "  📁 %s\n\n", m.rootDir)

	const maxVisible = 20
	start := 0
	if m.cursor >= maxVisible {
		start = m.cursor - maxVisible + 1
	}
	end := start + maxVisible
	if end > len(m.visible) {
		end = len(m.visible)
	}

	if len(m.visible) == 0 {
		colorMuted.Fprintln(b, "  (no .go files found)")
	}

	for i := start; i < end; i++ {
		n := m.tree[m.visible[i]]
		indent := strings.Repeat("  ", n.depth)
		isCursor := i == m.cursor

		var icon, nameStr string
		if n.isDir {
			if n.expanded {
				icon = colorWarn.Sprint("▼")
			} else {
				icon = colorWarn.Sprint("▶")
			}
			nameStr = colorWarn.Sprint(n.name + "/")
		} else if m.selected[n.fullPath] {
			icon = colorSuccess.Sprint("◆")
			nameStr = colorSuccess.Sprint(n.name)
		} else {
			icon = colorMuted.Sprint("◇")
			nameStr = colorInfo.Sprint(n.name)
		}

		if isCursor {
			colorActive.Fprintf(b, "  %s▸ %s %s\n", indent, icon, nameStr)
		} else {
			fmt.Fprintf(b, "  %s  %s %s\n", indent, icon, nameStr)
		}
	}

	if len(m.visible) > maxVisible {
		colorMuted.Fprintf(b, "\n  ... %d more\n", len(m.visible)-maxVisible)
	}

	b.WriteString("\n")
	colorMuted.Fprintln(b, "  ─────────────────────────────────────────────────")
	if cnt := len(m.selected); cnt > 0 {
		colorSuccess.Fprintf(b, "  ◆ %d selected", cnt)
		colorMuted.Fprintf(b, "   │  Tab to confirm\n")
	} else {
		colorMuted.Fprintln(b, "  No files selected")
	}
	colorMuted.Fprintf(b, "\n  ↑↓ move   → expand / ← collapse   Space select   a all   Tab confirm   q quit\n")
}

func (m TUIModel) renderConfirm(b *strings.Builder) {
	colorEmphasis.Fprintf(b, "  [5/5] Confirm & Run\n\n")
	if len(m.selected) == 0 {
		colorWarn.Fprintln(b, "  ⚠ No files selected — press B to go back.")
		return
	}
	colorSuccess.Fprintf(b, "  ✓ %d file(s) from %s\n\n", len(m.selected), filepath.Base(m.rootDir))

	files := collectGoFiles(m.tree)
	sort.Strings(files)
	limit := 16
	shown := 0
	for _, f := range files {
		if !m.selected[f] {
			continue
		}
		if shown >= limit {
			colorMuted.Fprintf(b, "  │  ... and %d more\n", len(m.selected)-limit)
			break
		}
		rel, _ := filepath.Rel(m.rootDir, f)
		colorMuted.Fprintf(b, "  │  ")
		colorInfo.Fprintln(b, rel)
		shown++
	}
	b.WriteString("\n")
	colorMuted.Fprintln(b, "  ─────────────────────────────────────────────────")
	colorMuted.Fprintln(b, "  ↵ Enter  run tests    B  go back    q  quit")
}

func RunTUI() (string, []string, string, string, error) {
	p := tea.NewProgram(InitialModel(), tea.WithAltScreen())
	fm, err := p.Run()
	if err != nil {
		return "", nil, "", "", err
	}
	m := fm.(TUIModel)
	if !m.confirmed || len(m.selected) == 0 {
		return "", nil, "", "", nil
	}
	var files []string
	for f := range m.selected {
		files = append(files, f)
	}
	sort.Strings(files)
	return m.rootDir, files, m.provider, m.selectedModel, nil
}
