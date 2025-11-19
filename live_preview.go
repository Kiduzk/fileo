package main

import (
	"fmt"
	"os"
	"path/filepath"
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

	selectedTreeBgColor  = lipgloss.Color("57")
	focusedBorderColor   = lipgloss.Color("62")
	unfocusedBorderColor = lipgloss.Color("240")
	panelBaseStyle       = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				Padding(0, 1)

	placeholderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("238"))

	endOfBufferStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("235"))

	focusedPlaceholderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("99"))

	focusedBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.HiddenBorder())

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
	children []treeItem
}

type model struct {
	width        int
	height       int
	keymap       keymap
	help         help.Model
	cfg          textarea.Model // left panel: command / yaml config
	treeItems    []treeItem     // flattened tree for navigation
	treeItemRoot treeItem
	cursor       int // selected item in tree
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

	// initial tree from cwd
	if wd, err := os.Getwd(); err == nil {
		m.rootPath = wd
		m.buildTree()
	}

	// Apply the config to the current directory and save it as map
	// The map has directory and its keys are what are subdirector

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
	helpView = lipgloss.PlaceHorizontal(m.width, lipgloss.Center, helpView)

	leftWidth := m.leftWidth
	previewWidth := m.width - leftWidth - 4
	panelHeight := m.height - helpHeight - 1

	leftBorderColor := unfocusedBorderColor
	if m.focusedPane == 0 {
		leftBorderColor = focusedBorderColor
	}
	leftStyle := panelBaseStyle.
		Width(leftWidth).
		Height(panelHeight).
		BorderForeground(leftBorderColor)

	leftPanel := leftStyle.Render(m.cfg.View())

	rightBorderColor := unfocusedBorderColor
	if m.focusedPane == 1 {
		rightBorderColor = focusedBorderColor
	}
	rightStyle := panelBaseStyle.
		Width(previewWidth).
		Height(panelHeight).
		BorderForeground(rightBorderColor). 
		Padding(1). 
		PaddingLeft(2)

	treeWidth := previewWidth - 2
	treeHeight := panelHeight - 2

	treeView := m.renderTree(treeWidth, treeHeight)
	rightPanel := rightStyle.Render(treeView)

	body := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)

	header := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Width(leftWidth).Align(lipgloss.Left).PaddingLeft(1).Render("Config text"),
		lipgloss.NewStyle().Width(previewWidth).Align(lipgloss.Left).PaddingLeft(3).Render("Directory Preview"),
	)

	return header + "\n" + body + "\n" + helpView
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

	// Initialize treeItemRoot
	m.treeItemRoot = m.treeItems[0]

	// Only recurse if root is expanded
	if rootExpanded {

		// First, we build the tree using destination paths
		for _, destPath := range ApplyConfigPreview([]byte(m.cfg.Value())) {
			m.buildTreeRecursive(destPath)
		}

		// Then we populate tree items accordingly
		m.populateTreeUIRecursive(m.treeItemRoot)
	}

	// Ensure cursor is in bounds
	if m.cursor >= len(m.treeItems) {
		m.cursor = len(m.treeItems) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

// Given a string path we decompose it into its constituents using the path separator and then
// add a tree item in to our directory
func (m *model) buildTreeRecursive(path string) {

	// Convert path separators to forward slashes for consistent splitting
	path = filepath.ToSlash(path)
	paths := strings.Split(path, "/")

	parentItem := &m.treeItemRoot
	for depth, childPath := range paths {
		if childPath == "" {
			continue
		}

		childFullPath := filepath.Join(parentItem.path, childPath)

		// Check if child already exists
		var existingChild *treeItem
		for i := range parentItem.children {
			if parentItem.children[i].name == childPath {
				existingChild = &parentItem.children[i]
				break
			}
		}

		if existingChild != nil {
			parentItem = existingChild
		} else {
			// For preview mode, check if file exists. If not, assume it's a directory
			// unless it's the last segment (which would be the file)
			isLastSegment := depth == len(paths)-1
			info, _ := os.Stat(childFullPath)

			var isDir bool
			if info != nil {
				isDir = info.IsDir()
			} else {
				// Virtual path (doesn't exist yet) - assume directories for all but last segment
				isDir = !isLastSegment
			}

			childItem := treeItem{
				name:     childPath,
				path:     childFullPath,
				isDir:    isDir,
				expanded: m.expandedDirs[childFullPath],
				children: []treeItem{},
				depth:    depth + 1,
			}

			parentItem.children = append(parentItem.children, childItem)
			parentItem = &parentItem.children[len(parentItem.children)-1]
		}
	}

}

// We populate the flat tree directory using the generated tree. This is called after tree is built
func (m *model) populateTreeUIRecursive(item treeItem) {
	for _, child := range item.children {

		// Check if child path already exists in treeItems
		found := false
		for _, existing := range m.treeItems {
			if existing.path == child.path {
				found = true
				break
			}
		}

		if !found {
			m.treeItems = append(m.treeItems, child)
		}

		if child.isDir && child.expanded {
			m.populateTreeUIRecursive(child)
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
		if width > 3 && len(line) > width {
			line = line[:width-3] + "..."
		} else if width <= 3 && len(line) > width {
			if width < 1 {
				width = 1
			}
			line = line[:width]
		}

		// Highlight cursor only when right pane is focused
		if i == m.cursor && m.focusedPane == 1 {
			style := lipgloss.NewStyle().
				Background(selectedTreeBgColor).
				Foreground(lipgloss.Color("230")).
				Bold(true).
				Width(width)
			line = style.Render(line)
		}

		b.WriteString(line + "\n")
	}

	return b.String()
}
