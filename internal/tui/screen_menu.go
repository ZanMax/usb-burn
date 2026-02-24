package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type menuItem struct {
	title string
	desc  string
	mode  Mode
}

type menuModel struct {
	items    []menuItem
	cursor   int
	selected bool
}

func newMenuModel() menuModel {
	return menuModel{
		items: []menuItem{
			{title: "Write Image to USB", desc: "Write an ISO/IMG/DMG/RAW file to a USB drive", mode: ModeWrite},
			{title: "Format USB Drive", desc: "Format a USB drive with a chosen filesystem", mode: ModeFormat},
		},
	}
}

func (m menuModel) Init() tea.Cmd {
	return nil
}

func (m menuModel) Update(msg tea.Msg) (menuModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, Keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, Keys.Down):
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case key.Matches(msg, Keys.Select):
			m.selected = true
		}
	}
	return m, nil
}

func (m menuModel) View() string {
	s := ""
	for i, item := range m.items {
		cursor := "  "
		style := NormalItemStyle
		descStyle := DescStyle

		if i == m.cursor {
			cursor = SelectedStyle.Render("> ")
			style = ActiveItemStyle
			descStyle = descStyle.Foreground(ColorYellow)
		}

		s += cursor + style.Render(item.title) + "\n"
		s += descStyle.Render(item.desc) + "\n"
		if i < len(m.items)-1 {
			s += "\n"
		}
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorOrange).
		Padding(1, 2).
		Render(s)

	return box
}
