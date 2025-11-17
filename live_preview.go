package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	helpHeight = 5
)

var (
	cursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))

	cursorLineStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("57")).
			Foreground(lipgloss.Color("230"))

	placeholderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("238"))

	endOfBufferStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("235"))

	focusedPlaceholderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("99"))

	focusedBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("238"))

	blurredBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.HiddenBorder())
)

type keymap = struct {
	switchPanel, refresh, quit, up, down, toggle key.Binding
}

func newTextarea() textarea.Model {
	t := textarea.New()
	t.Prompt = ""
	t.Placeholder = "Type something"
	t.ShowLineNumbers = true
	t.Cursor.Style = cursorStyle
	t.FocusedStyle.Placeholder = focusedPlaceholderStyle
	t.BlurredStyle.Placeholder = placeholderStyle
	t.FocusedStyle.CursorLine = cursorLineStyle
	t.FocusedStyle.Base = focusedBorderStyle
	t.BlurredStyle.Base = blurredBorderStyle
	t.FocusedStyle.EndOfBuffer = endOfBufferStyle
	t.BlurredStyle.EndOfBuffer = endOfBufferStyle
	t.KeyMap.DeleteWordBackward.SetEnabled(false)
	t.KeyMap.LineNext = key.NewBinding(key.WithKeys("down"))
	t.KeyMap.LinePrevious = key.NewBinding(key.WithKeys("up"))
	t.CharLimit = 0 
	t.Blur()
	return t
}

type treeItem struct {
	path     string
	name     string
	isDir    bool
	expanded bool
	depth    int
}

type model struct {
	width        int
	height       int
	keymap       keymap
	help         help.Model
	cfg          textarea.Model // left panel: command / yaml config
	treeItems    []treeItem     // flattened tree for navigation
	cursor       int            // selected item in tree
	leftWidth    int
	focusedPane  int // 0 = left (config), 1 = right (tree)
	rootPath     string
	expandedDirs map[string]bool // tracks which dirs are expanded
	cfgFilePath  string
}

func newModel(cfgFilePath string) model {
	m := model{
		cfgFilePath:  cfgFilePath,
		help:         help.New(),
		focusedPane:  0, // start with config panel focused
		expandedDirs: make(map[string]bool),
		keymap: keymap{
			switchPanel: key.NewBinding(
				key.WithKeys("tab"),
				key.WithHelp("tab", "switch panel"),
			),
			refresh: key.NewBinding(
				key.WithKeys("ctrl+r"),
				key.WithHelp("ctrl+r", "refresh"),
			),
			up: key.NewBinding(
				key.WithKeys("up", "k"),
				key.WithHelp("↑/k", "up"),
			),
			down: key.NewBinding(
				key.WithKeys("down", "j"),
				key.WithHelp("↓/j", "down"),
			),
			toggle: key.NewBinding(
				key.WithKeys("enter", " "),
				key.WithHelp("enter/space", "expand/collapse"),
			),
			quit: key.NewBinding(
				key.WithKeys("q", "ctrl+c"),
				key.WithHelp("q", "quit"),
			),
		},
	}

	// single config textarea on the left
	m.cfg = newTextarea()
	m.cfg.Placeholder = "Enter YAML config here"
	m.cfg.Focus()

	// initial tree from cwd
	if wd, err := os.Getwd(); err == nil {
		m.rootPath = wd
		m.buildTree()
	}

	m.width = 10

	// Read the config file and render that to the user
	data, err := os.ReadFile(m.cfgFilePath)
	if err != nil {
		m.cfg.SetValue("could not read config file, make sure it exists")
	} else {
		m.cfg.SetValue(string(data))

		// Temporary fix -- TODO: figure out a better way to do this
		for range 100 {
			m.cfg.CursorUp()
		}
	}

	return m
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {

		case key.Matches(msg, m.keymap.quit):
			return m, tea.Quit

		case key.Matches(msg, m.keymap.switchPanel):
			m.focusedPane = 1 - m.focusedPane // toggle 0<->1
			if m.focusedPane == 0 {
				m.cfg.Focus()
			} else {
				m.cfg.Blur()
			}
			return m, nil

		case key.Matches(msg, m.keymap.refresh):
			m.buildTree()
			return m, nil

		case key.Matches(msg, m.keymap.up):
			if m.focusedPane == 1 && len(m.treeItems) > 0 {
				m.cursor--
				if m.cursor < 0 {
					m.cursor = 0
				}
				return m, nil
			}
		case key.Matches(msg, m.keymap.down):
			if m.focusedPane == 1 && len(m.treeItems) > 0 {
				m.cursor++
				if m.cursor >= len(m.treeItems) {
					m.cursor = len(m.treeItems) - 1
				}
				return m, nil
			}
		case key.Matches(msg, m.keymap.toggle):
			if m.focusedPane == 1 && len(m.treeItems) > 0 && m.cursor < len(m.treeItems) {
				if m.treeItems[m.cursor].isDir {

					// Toggle the expansion state in persistent map
					path := m.treeItems[m.cursor].path
					m.expandedDirs[path] = !m.expandedDirs[path]

					// Rebuild tree to reflect the change
					m.buildTree()
				}
				return m, nil
			}
		}
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
	}

	m.sizeInputs()

	// Only update config textarea if left panel is focused
	if m.focusedPane == 0 {
		newCfg, cmd := m.cfg.Update(msg)
		m.cfg = newCfg
		cmds = append(cmds, cmd)
	}

	// Tehcnically, this just has atmost 1 element but makes it easy in the future if we want to batch commands
	return m, tea.Batch(cmds...)
}

func (m *model) sizeInputs() {
	left := m.width / 2
	m.leftWidth = left

	// Set textarea size (accounting for border and padding)
	textAreaWidth := left - 2
	textAreaHeight := m.height - helpHeight - 4

	m.cfg.SetWidth(textAreaWidth)
	m.cfg.SetHeight(textAreaHeight)
}

func (m model) View() string {
	helpView := m.help.ShortHelpView([]key.Binding{
		m.keymap.switchPanel,
		m.keymap.up,
		m.keymap.down,
		m.keymap.toggle,
		m.keymap.refresh,
		m.keymap.quit,
	})

	// Calculate dimensions
	leftWidth := m.leftWidth
	previewWidth := m.width - leftWidth - 4 // account for borders
	panelHeight := m.height - helpHeight - 2

	// Left panel with config textarea
	leftBorderColor := lipgloss.Color("240")
	if m.focusedPane == 0 {
		leftBorderColor = lipgloss.Color("62") // bright when focused
	}
	leftStyle := lipgloss.NewStyle().
		Width(leftWidth).
		Height(panelHeight).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(leftBorderColor).
		Padding(0, 1)

	leftPanel := leftStyle.Render(m.cfg.View())

	// Right panel with tree
	rightBorderColor := lipgloss.Color("240")
	if m.focusedPane == 1 {
		rightBorderColor = lipgloss.Color("62") // bright when focused
	}
	rightStyle := lipgloss.NewStyle().
		Width(previewWidth).
		Height(panelHeight).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(rightBorderColor).
		Padding(0, 1)

	treeView := m.renderTree(previewWidth-4, panelHeight-2)
	rightPanel := rightStyle.Render(treeView)

	body := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)

	return body + "\n" + helpView
}

func RunLivePreview(previewConfig string) {
	if _, err := tea.NewProgram(newModel(previewConfig), tea.WithAltScreen()).Run(); err != nil {
		// When using alternate screen, print to stderr to ensure visibility.
		fmt.Fprintln(os.Stderr, "Error while running program:", err)
		os.Exit(1)
	}
}

// buildTree constructs a flattened tree of items respecting expand/collapse state
func (m *model) buildTree() {
	if m.rootPath == "" {
		return
	}

	// Clear and rebuild
	m.treeItems = []treeItem{}

	// Root item - check expanded state, default to true on first load
	rootExpanded, exists := m.expandedDirs[m.rootPath]
	if !exists {
		rootExpanded = true
		m.expandedDirs[m.rootPath] = true
	}

	m.treeItems = append(m.treeItems, treeItem{
		path:     m.rootPath,
		name:     filepath.Base(m.rootPath),
		isDir:    true,
		expanded: rootExpanded,
		depth:    0,
	})

	// Only recurse if root is expanded
	if rootExpanded {
		m.buildTreeRecursive(m.rootPath, 0)
	}

	// Ensure cursor is in bounds
	if m.cursor >= len(m.treeItems) {
		m.cursor = len(m.treeItems) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

func (m *model) buildTreeRecursive(path string, depth int) {
	// Don't go too deep
	if depth >= 5 {
		return
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return
	}

	// Sort entries
	sort.Slice(entries, func(i, j int) bool {
		// Directories first, then alphabetically
		if entries[i].IsDir() != entries[j].IsDir() {
			return entries[i].IsDir()
		}
		return entries[i].Name() < entries[j].Name()
	})

	for _, e := range entries {
		fullPath := filepath.Join(path, e.Name())
		isExpanded := m.expandedDirs[fullPath]

		item := treeItem{
			path:     fullPath,
			name:     e.Name(),
			isDir:    e.IsDir(),
			expanded: isExpanded,
			depth:    depth + 1,
		}
		m.treeItems = append(m.treeItems, item)

		// Recurse if directory and expanded
		if e.IsDir() && isExpanded {
			m.buildTreeRecursive(fullPath, depth+1)
		}
	}
}

// renderTree renders the tree with cursor highlight
func (m model) renderTree(width, height int) string {
	var b strings.Builder

	visibleStart := 0
	visibleEnd := len(m.treeItems)

	// Calculate scroll window if tree is too tall
	if len(m.treeItems) > height {
		visibleStart = m.cursor - height/2
		if visibleStart < 0 {
			visibleStart = 0
		}
		visibleEnd = visibleStart + height
		if visibleEnd > len(m.treeItems) {
			visibleEnd = len(m.treeItems)
			visibleStart = visibleEnd - height
			if visibleStart < 0 {
				visibleStart = 0
			}
		}
	}

	for i := visibleStart; i < visibleEnd; i++ {
		item := m.treeItems[i]
		indent := strings.Repeat("  ", item.depth)

		var icon string
		if item.isDir {
			if item.expanded {
				icon = "▾ "
			} else {
				icon = "▸ "
			}
		} else {
			icon = "  "
		}

		line := indent + icon + item.name
		if item.isDir {
			line += "/"
		}

		// Truncate if too long (before styling)
		if len(line) > width {
			line = line[:width-3] + "..."
		}

		// Highlight cursor only when right pane is focused
		if i == m.cursor && m.focusedPane == 1 {
			style := lipgloss.NewStyle().
				Background(lipgloss.Color("62")).
				Foreground(lipgloss.Color("230")).
				Bold(true).
				Width(width)
			line = style.Render(line)
		}

		b.WriteString(line + "\n")
	}

	return b.String()
}
