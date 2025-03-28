package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	helpHeight    = 5
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
	changeTab, quit key.Binding
}

func newTextarea() textarea.Model {
	t := textarea.New()
  t.SetValue(`folders:
  - name: "DocumentsSample"
    recurse: True
    extensions: 
      - "go"
      - "mod"
    patterns:
      - "READ"
  `)
	t.Prompt = ""
	t.Placeholder = "Type the config here "
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
	t.Blur()
	return t
}

type model struct {
	width  int
	height int
	keymap keymap
	help   help.Model
	input textarea.Model
  filePicker Model 
  isInputFocus bool
}

func newModel(fp Model) model {
	m := model{
		help:   help.New(),
		keymap: keymap{
      changeTab: key.NewBinding(
        key.WithKeys("tab"),
        key.WithHelp("tab", "change tabs"),
      ),
			quit: key.NewBinding(
				key.WithKeys("esc", "ctrl+c"),
				key.WithHelp("esc", "quit"),
			),
		},
    filePicker: fp,
	}

  m.input = newTextarea()
  m.filePicker = New() 
	m.input.Focus()


	// m.updateKeybindings()
	return m
}

func (m model) Init() tea.Cmd {
  return m.filePicker.Init()
	// return textarea.Blink
 }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

  if !m.isInputFocus {
    m.filePicker, _ = m.filePicker.Update(msg)
  }

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.quit):
      m.input.Blur()
			return m, tea.Quit
		case key.Matches(msg, m.keymap.changeTab):
      m.isInputFocus = !m.isInputFocus

      if !m.isInputFocus {
        m.input.Blur()
      } else {
        m.input.Focus()
        cmd := m.input.Focus()
        cmds = append(cmds, cmd)
      }
		}
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
	}

	// m.updateKeybindings()
	m.sizeInputs()

  // Update the views
  m.filePicker.Update(msg)
	newModel, cmd := m.input.Update(msg)
  m.input = newModel
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *model) sizeInputs() {
  m.input.SetWidth(m.width / 2)
  m.input.SetHeight(m.height - helpHeight)

  // m.filePicker.Width = m.width / 2
  m.filePicker.SetHeight(m.height - helpHeight)
}

// func (m *model) updateKeybindings() {
// 	m.keymap.add.SetEnabled(len(m.inputs) < maxInputs)
// 	m.keymap.remove.SetEnabled(len(m.inputs) > minInputs)
// }

func (m model) View() string {
	help := m.help.ShortHelpView([]key.Binding{
    m.keymap.changeTab,
		m.keymap.quit,
	})

	var views []string

  views = append(views, m.input.View())
  views = append(views, m.filePicker.View())

	return lipgloss.JoinHorizontal(lipgloss.Center, views...) + "\n\n" + help
}

func maintui() {
  fp := New()
  tempDir, err := os.MkdirTemp("", "this")
  defer os.RemoveAll(tempDir)


  // Create the mock file system here 
  os.WriteFile(tempDir+"/mockfile.txt", []byte{}, 0644)
  os.WriteFile(tempDir+"/mockfile1.txt", []byte{}, 0644)

  os.MkdirAll(filepath.Join(tempDir, "folder1/asfd"), 0644)

  HandleError(err)

  fp.CurrentDirectory = tempDir 
  // fp.AllowedTypes = []string{".mod", ".sum", ".go", ".txt", ".md"}


	if _, err := tea.NewProgram(newModel(fp), tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error while running program:", err)
		os.Exit(1)
	}
}


