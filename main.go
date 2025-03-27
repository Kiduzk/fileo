package main

import (
	"fmt"
	"os"
  "testing/fstest"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/filepicker"
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
	next, prev, add, remove, quit key.Binding
}

func newTextarea() textarea.Model {
	t := textarea.New()
	t.Prompt = ""
	t.Placeholder = "Type something"
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
	inputs []textarea.Model
	focus  int
  filePicker filepicker.Model 
}

func newModel(fp filepicker.Model) model {
	m := model{
		inputs: make([]textarea.Model, 2),
		help:   help.New(),
		keymap: keymap{
			next: key.NewBinding(
				key.WithKeys("tab"),
				key.WithHelp("tab", "next"),
			),
			prev: key.NewBinding(
				key.WithKeys("shift+tab"),
				key.WithHelp("shift+tab", "prev"),
			),
			add: key.NewBinding(
				key.WithKeys("ctrl+n"),
				key.WithHelp("ctrl+n", "add an editor"),
			),
			remove: key.NewBinding(
				key.WithKeys("ctrl+w"),
				key.WithHelp("ctrl+w", "remove an editor"),
			),
			quit: key.NewBinding(
				key.WithKeys("esc", "ctrl+c"),
				key.WithHelp("esc", "quit"),
			),
		},
    filePicker: fp,
	}

  m.inputs[0] = newTextarea()
  m.inputs[1] = newTextarea()
	m.inputs[m.focus].Focus()


	// m.updateKeybindings()
	return m
}

func (m model) Init() tea.Cmd {
  return m.filePicker.Init()
	// return textarea.Blink
 }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

  m.filePicker, _ = m.filePicker.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.quit):
			for i := range m.inputs {
				m.inputs[i].Blur()
			}
			return m, tea.Quit
		case key.Matches(msg, m.keymap.next):
			m.inputs[m.focus].Blur()
			m.focus++
			if m.focus > len(m.inputs)-1 {
				m.focus = 0
			}
			cmd := m.inputs[m.focus].Focus()
			cmds = append(cmds, cmd)
		case key.Matches(msg, m.keymap.prev):
			m.inputs[m.focus].Blur()
			m.focus--
			if m.focus < 0 {
				m.focus = len(m.inputs) - 1
			}
			cmd := m.inputs[m.focus].Focus()
			cmds = append(cmds, cmd)
		// case key.Matches(msg, m.keymap.add):
		// 	m.inputs = append(m.inputs, newTextarea())
		case key.Matches(msg, m.keymap.remove):
			m.inputs = m.inputs[:len(m.inputs)-1]
			if m.focus > len(m.inputs)-1 {
				m.focus = len(m.inputs) - 1
			}
		}
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
	}

	// m.updateKeybindings()
	m.sizeInputs()

	// Update all textareas
	for i := range m.inputs {
		newModel, cmd := m.inputs[i].Update(msg)
		m.inputs[i] = newModel
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *model) sizeInputs() {
	for i := range m.inputs {
		m.inputs[i].SetWidth(m.width / len(m.inputs))
		m.inputs[i].SetHeight(m.height - helpHeight)
	}
}

// func (m *model) updateKeybindings() {
// 	m.keymap.add.SetEnabled(len(m.inputs) < maxInputs)
// 	m.keymap.remove.SetEnabled(len(m.inputs) > minInputs)
// }

func (m model) View() string {
	help := m.help.ShortHelpView([]key.Binding{
		m.keymap.next,
		m.keymap.prev,
		m.keymap.add,
		m.keymap.remove,
		m.keymap.quit,
	})

	var views []string
	// for i := range m.inputs {
	// 	views = append(views, m.inputs[i].View())
	// }

  views = append(views, m.inputs[0].View())
  views = append(views, m.filePicker.View())

	return lipgloss.JoinHorizontal(lipgloss.Center, views...) + "\n\n" + help
}

func main() {
  fp := filepicker.New()
  fp.CurrentDirectory, _ = os.UserHomeDir()
  // fp.AllowedTypes = []string{".mod", ".sum", ".go", ".txt", ".md"}


	if _, err := tea.NewProgram(newModel(fp), tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error while running program:", err)
		os.Exit(1)
	}
}




// package main
//
// import (
// 	// "fmt"
// 	"log"
// 	"os"
//
// 	"github.com/urfave/cli/v2"
// )
//
// func main() {
//   ApplyConfig("config.yaml")
//   return 
//   app := &cli.App{
//     Flags: []cli.Flag{
//       &cli.StringFlag{
//         Name: "output",
//         Usage: "The directory of output files.",
//         Aliases: []string{"o"},
//       },
//       &cli.StringSliceFlag{
//         Name: "extension",
//         Usage: "Matches files with a specific extension.",
//         Aliases: []string{"e"},
//       },
//       &cli.StringSliceFlag{
//         Name: "pattern",
//         Usage: "Pattern to match with file name, supports regex.",
//         Aliases: []string{"p"},
//       },
//       &cli.BoolFlag{
//         Name: "recursive",
//         Usage: "Option to recursively search a directory.",
//         Aliases: []string{"r"},
//       },
//     },
//     Name: "FileOrganizer",
//     Usage: "Organizes files nicely.",
//     Action: cliActionHandler,
//   }
//   if err := app.Run(os.Args); err != nil {
//     log.Fatal(err)
//   }
// }
//
//
// func cliActionHandler(cCtx *cli.Context) error {
//   if cCtx.NArg() > 0 {
//     // pattern := cCtx.Args().Get(0)
//   }
//
//   // Get some of the cli arguments
//   outputPath := cCtx.String("output")
//
//   patternSlice := cCtx.StringSlice("pattern")
//   extensionSlice := cCtx.StringSlice("extension")
//
//   mimeType := cCtx.String("mime")
//
//   recursive := cCtx.Bool("recursive")
//
//   if outputPath == "" {
//     log.Fatal("No file output path given.")
//     return nil
//   }
//
//   if len(patternSlice) != 0 {
//     var organizeFunction func(string, string) 
//
//     if recursive {
//       organizeFunction = OrganizeFilesByRegexRecursive
//     } else {
//       organizeFunction = OrganizeFilesByRegex
//     }
//
//
//     for _, pattern := range patternSlice {
//       organizeFunction(string(pattern), outputPath)
//     }
//
//   } else if (len(extensionSlice) != 0) {
//
//     var organizeFunction func(string, string) 
//
//     if recursive {
//       organizeFunction = OrganizeFilesByExtension
//     } else {
//       organizeFunction = OrganizeFilesByExtensionRecursive
//     }
//
//     for _, extension := range extensionSlice {
//       organizeFunction(outputPath, string(extension))
//     }
//
//   } else if (mimeType != "") {
//
//   } 
//
//   return nil
// }

