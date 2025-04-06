package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/filepicker"
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
	leftFilePicker   filepicker.Model
	rightFilePicker   filepicker.Model
	isLeftFocus bool
}

func newModel(rightFP, leftFP filepicker.Model) model {
	m := model{
		isLeftFocus: false,
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

	// m.updateKeybindings()
	return m
}

func (m model) Init() tea.Cmd {
	var cmds []tea.Cmd
	cmds = append(cmds, m.leftFilePicker.Init())
	cmds = append(cmds, m.rightFilePicker.Init())

	return tea.Batch(cmds...)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.quit):
			return m, tea.Quit
		case key.Matches(msg, m.keymap.changeTab):
			m.isLeftFocus = !m.isLeftFocus

		// TODO: some sort of blur here would be nice

			// if !m.isLeftFocus {
			// 	m.input.Blur()
			// } else {
			// 	m.input.Focus()
			// 	cmd := m.input.Focus()
			// 	cmds = append(cmds, cmd)
			// }
		}
	}

	if m.isLeftFocus {
		m.leftFilePicker, cmd = m.leftFilePicker.Update(msg)
	} else {
		m.rightFilePicker, cmd = m.rightFilePicker.Update(msg)
	}

	return m, cmd
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
	fpTemp := filepicker.New()
	fpTemp2 := filepicker.New()

	fpTemp2.CurrentDirectory, _ = os.UserHomeDir()


	tempDir, err := os.MkdirTemp("", "this")
	defer os.RemoveAll(tempDir)

	// Create the mock file system here
	os.WriteFile(tempDir+"/mockfile.txt", []byte{}, 0644)
	os.WriteFile(tempDir+"/mockfile1.txt", []byte{}, 0644)

	os.MkdirAll(filepath.Join(tempDir, "folder1/asfd"), 0644)

	HandleError(err)

	fpTemp.CurrentDirectory = tempDir
	// fp.AllowedTypes = []string{".mod", ".sum", ".go", ".txt", ".md"}

	if _, err := tea.NewProgram(newModel(fpTemp, fpTemp2), tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error while running program:", err)
		os.Exit(1)
	}
}
