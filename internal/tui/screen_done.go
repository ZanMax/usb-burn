package tui

import (
	"fmt"

	"usb_burn/internal/device"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type doneModel struct {
	restarted bool
	quit      bool
}

func newDoneModel() doneModel {
	return doneModel{}
}

func (m doneModel) Init() tea.Cmd {
	return nil
}

func (m doneModel) Update(msg tea.Msg) (doneModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, Keys.Restart):
			m.restarted = true
		case key.Matches(msg, Keys.Quit), key.Matches(msg, Keys.Select):
			m.quit = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m doneModel) ViewSuccess(state *AppState) string {
	s := ""
	if state.Mode == ModeWrite {
		s += lipgloss.NewStyle().Foreground(ColorGreen).Bold(true).Render("Image written successfully!") + "\n\n"
		if state.BytesWritten > 0 {
			s += fmt.Sprintf("  Bytes written: %s\n", device.FormatBytes(state.BytesWritten))
		}
		if state.Duration > 0 {
			s += fmt.Sprintf("  Duration:      %s\n", formatDuration(state.Duration))
			if state.BytesWritten > 0 {
				speed := float64(state.BytesWritten) / state.Duration
				s += fmt.Sprintf("  Avg speed:     %s/s\n", device.FormatBytes(int64(speed)))
			}
		}
	} else {
		s += lipgloss.NewStyle().Foreground(ColorGreen).Bold(true).Render("Drive formatted successfully!") + "\n\n"
		s += fmt.Sprintf("  Format: %s\n", state.FormatLabel)
		s += fmt.Sprintf("  Drive:  %s\n", state.SelectedDrive.DisplayName())
	}

	s += "\n" + DimStyle.Render("Press r to start over, q or enter to quit")

	return SuccessBoxStyle.Render(s)
}

func (m doneModel) ViewError(err error) string {
	s := ErrorStyle.Render("Operation failed") + "\n\n"
	s += fmt.Sprintf("  %s\n", err.Error())
	s += "\n" + DimStyle.Render("Press r to start over, q or enter to quit")

	return DangerBoxStyle.Render(s)
}
