package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
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
	changeTab, quit key.Binding
}

type model struct {
	width        int
	height       int
	keymap       keymap
	help         help.Model
	leftFilePicker   Model
	rightFilePicker   Model
  started bool
}

func newModel(rightFP, leftFP Model) model {
  leftFP.isBlurred = false 
  rightFP.isBlurred = true

	m := model{
		help: help.New(),
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
		leftFilePicker: leftFP,
		rightFilePicker: rightFP,
	}

	return m
}

func (m model) Init() tea.Cmd {
	var cmds []tea.Cmd

	cmds = append(cmds, m.leftFilePicker.Init())
	cmds = append(cmds, m.rightFilePicker.Init())

	return tea.Batch(cmds...)
}


func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
  var cmd tea.Cmd


  switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.quit):
			return m, tea.Quit
		case key.Matches(msg, m.keymap.changeTab):

      if m.leftFilePicker.isBlurred {
        m.leftFilePicker.isBlurred = false 
        m.rightFilePicker.isBlurred = true 
      } else {
        m.rightFilePicker.isBlurred = false
        m.leftFilePicker.isBlurred = true 
      }
		}
	}

  m.leftFilePicker, cmd = m.leftFilePicker.Update(msg)
  cmds = append(cmds, cmd)
  m.rightFilePicker, cmd = m.rightFilePicker.Update(msg)
  cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...) 
}


func (m model) View() string {
	help := m.help.ShortHelpView([]key.Binding{
		m.keymap.changeTab,
		m.keymap.quit,
	})

	var views []string

	left := m.leftFilePicker.View()
	right := m.rightFilePicker.View()


	views = append(views, left)
	views = append(views, right)

	return lipgloss.JoinHorizontal(lipgloss.Center, views...) + "\n\n" + help
}


// Takes in the name of the config file and creates a live preview based on that
func RunLivePreview(configFile string) {
	leftFP := New()
  rightFP := New()

	// fpTemp2.CurrentDirectory, _ = os.UserHomeDir()

	tempDir, err := os.MkdirTemp("", "this")
	defer os.RemoveAll(tempDir)

	// Create the mock file system here
	os.WriteFile(tempDir+"/mockfile.txt", []byte{}, 0644)
	os.WriteFile(tempDir+"/mockfile1.txt", []byte{}, 0644)

	os.MkdirAll(filepath.Join(tempDir, "folder1/asfd"), 0644)

	HandleError(err)

	rightFP.CurrentDirectory = tempDir
	// fp.AllowedTypes = []string{".mod", ".sum", ".go", ".txt", ".md"}

	if _, err := tea.NewProgram(newModel(leftFP, rightFP), tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error while running program:", err)
		os.Exit(1)
	}
}
