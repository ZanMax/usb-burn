package tui

import (
	"os"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type dirPickerModel struct {
	textInput textinput.Model
	err       string
	confirmed bool
}

func newDirPickerModel(defaultDir string) dirPickerModel {
	ti := textinput.New()
	ti.Placeholder = "Enter directory path"
	ti.SetValue(defaultDir)
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50
	ti.PromptStyle = lipgloss.NewStyle().Foreground(ColorOrange)
	ti.TextStyle = lipgloss.NewStyle().Foreground(ColorWhite)

	return dirPickerModel{
		textInput: ti,
	}
}

func (m dirPickerModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m dirPickerModel) Update(msg tea.Msg) (dirPickerModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, Keys.Select):
			// Validate directory
			dir := m.textInput.Value()
			info, err := os.Stat(dir)
			if err != nil {
				m.err = "Directory does not exist"
				return m, nil
			}
			if !info.IsDir() {
				m.err = "Path is not a directory"
				return m, nil
			}
			m.err = ""
			m.confirmed = true
			return m, nil
		default:
			m.err = ""
		}
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m dirPickerModel) View() string {
	s := DimStyle.Render("Enter the directory to scan for image files:") + "\n\n"
	s += m.textInput.View() + "\n"

	if m.err != "" {
		s += "\n" + ErrorStyle.Render(m.err)
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorOrange).
		Padding(1, 2).
		Render(s)

	return box
}
