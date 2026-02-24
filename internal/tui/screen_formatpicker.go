package tui

import (
	"usb_burn/internal/formatter"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type formatPickerModel struct {
	formats  []formatter.FSType
	cursor   int
	selected bool
}

func newFormatPickerModel() formatPickerModel {
	return formatPickerModel{
		formats: formatter.AvailableFormats(),
	}
}

func (m formatPickerModel) Init() tea.Cmd {
	return nil
}

func (m formatPickerModel) Update(msg tea.Msg) (formatPickerModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, Keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, Keys.Down):
			if m.cursor < len(m.formats)-1 {
				m.cursor++
			}
		case key.Matches(msg, Keys.Select):
			if len(m.formats) > 0 {
				m.selected = true
			}
		}
	}
	return m, nil
}

func (m formatPickerModel) View() string {
	if len(m.formats) == 0 {
		return BoxStyle.Render(ErrorStyle.Render("No supported filesystem formats available on this system"))
	}

	s := DimStyle.Render("Select filesystem format:") + "\n\n"

	for i, fs := range m.formats {
		cursor := "  "
		nameStyle := NormalItemStyle
		descStyle := DescStyle

		if i == m.cursor {
			cursor = SelectedStyle.Render("> ")
			nameStyle = ActiveItemStyle
			descStyle = lipgloss.NewStyle().Foreground(ColorYellow).PaddingLeft(4)
		}

		s += cursor + nameStyle.Render(fs.Name) + "\n"
		s += descStyle.Render(fs.Description) + "\n"
		if i < len(m.formats)-1 {
			s += "\n"
		}
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorOrange).
		Padding(1, 2).
		Render(s)
}
