package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
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
	modeInput   = "input"
	modeBrowse  = "browse"
	modeConfirm = "confirm"
)

// ── Input field with real cursor support ──────────────────────────────────────

type inputField struct {
	value  string
	cursor int // byte offset into value
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

// view renders the field with the char under cursor highlighted.
func (f inputField) view() string {
	before := f.value[:f.cursor]
	after := f.value[f.cursor:]
	if len(after) == 0 {
		return colorInfo.Sprint(before) + colorActive.Sprint("█")
	}
	_, sz := utf8.DecodeRuneInString(after)
	return colorInfo.Sprint(before) + colorActive.Sprint(after[:sz]) + colorInfo.Sprint(after[sz:])
}

// ── Tree node ─────────────────────────────────────────────────────────────────

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

// buildVisible returns the indices of nodes that should be shown.
// A node is hidden when any ancestor dir is collapsed.
func buildVisible(nodes []treeNode) []int {
	var visible []int
	collapsedDepth := -1
	for i, n := range nodes {
		if collapsedDepth >= 0 {
			if n.depth > collapsedDepth {
				continue // hidden under a collapsed dir
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

// ── Model ─────────────────────────────────────────────────────────────────────

type TUIModel struct {
	mode      string
	field     inputField
	rootDir   string
	tree      []treeNode
	visible   []int
	cursor    int
	selected  map[string]bool
	confirmed bool
	err       error
}

func InitialModel() TUIModel {
	cwd, _ := os.Getwd()
	return TUIModel{
		mode:     modeInput,
		field:    newField(cwd),
		selected: make(map[string]bool),
	}
}

func (m TUIModel) Init() tea.Cmd { return nil }

// ── Update ────────────────────────────────────────────────────────────────────

func (m TUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	k, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch m.mode {
	case modeInput:
		return m.handleInput(k)
	case modeBrowse:
		return m.handleBrowse(k)
	case modeConfirm:
		return m.handleConfirm(k)
	}
	return m, nil
}

func (m TUIModel) handleInput(k tea.KeyMsg) (TUIModel, tea.Cmd) {
	switch k.String() {
	case "ctrl+c", "esc":
		return m, tea.Quit
	case "enter":
		path := strings.TrimSpace(m.field.value)
		if path == "" {
			path = "."
		}
		abs, err := filepath.Abs(path)
		if err != nil {
			m.err = fmt.Errorf("invalid path: %s", path)
			return m, nil
		}
		if info, err := os.Stat(abs); err != nil || !info.IsDir() {
			m.err = fmt.Errorf("not a directory: %s", abs)
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
		if len(k.Runes) > 0 {
			m.field.insert(k.Runes[0])
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
			// Keep cursor on same node
			for i, vi := range m.visible {
				if vi == idx {
					m.cursor = i
					break
				}
			}
		} else if !n.isDir {
			// Toggle file selection
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
			// Collapse this dir
			m.tree[idx].expanded = false
			m.visible = buildVisible(m.tree)
			for i, vi := range m.visible {
				if vi == idx {
					m.cursor = i
					break
				}
			}
		} else {
			// Jump to parent dir
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

// ── Helpers ───────────────────────────────────────────────────────────────────

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

// ── View ──────────────────────────────────────────────────────────────────────

func (m TUIModel) View() string {
	var b strings.Builder
	// Banner
	b.WriteString("\n")
	for _, line := range verdictArt {
		colorPrimary.Fprintf(&b, "  %s\n", line)
	}
	colorMuted.Fprintf(&b, "  %s\n", strings.Repeat("─", 55))
	colorMuted.Fprintf(&b, "  AI-powered Go test generation  ◆  v0.1\n\n")

	switch m.mode {
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

func (m TUIModel) renderInput(b *strings.Builder) {
	colorEmphasis.Fprintf(b, "  [1/3] Enter directory path\n\n")
	colorMuted.Fprintf(b, "  ▶ ")
	b.WriteString(m.field.view())
	b.WriteString("\n\n")
	colorMuted.Fprintln(b, "  ← →  move cursor    Home/End  jump    ↵ Enter  open    ESC  quit")
}

func (m TUIModel) renderBrowse(b *strings.Builder) {
	colorEmphasis.Fprintf(b, "  [2/3] Select files\n\n")
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
	colorEmphasis.Fprintf(b, "  [3/3] Confirm & Run\n\n")
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

// ── Public API ────────────────────────────────────────────────────────────────

func RunTUI() (string, []string, error) {
	p := tea.NewProgram(InitialModel(), tea.WithAltScreen())
	fm, err := p.Run()
	if err != nil {
		return "", nil, err
	}
	m := fm.(TUIModel)
	if !m.confirmed || len(m.selected) == 0 {
		return "", nil, nil
	}
	var files []string
	for f := range m.selected {
		files = append(files, f)
	}
	sort.Strings(files)
	return m.rootDir, files, nil
}
