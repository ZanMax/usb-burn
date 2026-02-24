package tui

import (
	"fmt"

	"usb_burn/internal/device"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type drivePickerModel struct {
	drives   []device.Drive
	cursor   int
	loading  bool
	err      string
	selected bool
	spinner  spinner.Model
}

func newDrivePickerModel() drivePickerModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(ColorOrange)

	return drivePickerModel{
		loading: true,
		spinner: s,
	}
}

func (m drivePickerModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, detectDrives())
}

func detectDrives() tea.Cmd {
	return func() tea.Msg {
		drives, err := device.DetectDrives()
		return drivesDetectedMsg{Drives: drives, Err: err}
	}
}

func (m drivePickerModel) Update(msg tea.Msg) (drivePickerModel, tea.Cmd) {
	switch msg := msg.(type) {
	case drivesDetectedMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err.Error()
			return m, nil
		}
		m.drives = msg.Drives
		if len(m.drives) == 0 {
			m.err = "No USB drives detected"
		} else {
			m.err = ""
		}
		m.cursor = 0
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, Keys.Up):
			if !m.loading && m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, Keys.Down):
			if !m.loading && m.cursor < len(m.drives)-1 {
				m.cursor++
			}
		case key.Matches(msg, Keys.Select):
			if !m.loading && len(m.drives) > 0 {
				m.selected = true
			}
		case key.Matches(msg, Keys.Refresh):
			m.loading = true
			m.err = ""
			m.drives = nil
			return m, tea.Batch(m.spinner.Tick, detectDrives())
		}
	}
	return m, nil
}

func (m drivePickerModel) View() string {
	if m.loading {
		return BoxStyle.Render(m.spinner.View() + " Detecting USB drives...")
	}

	if m.err != "" {
		content := ErrorStyle.Render(m.err) + "\n\n" +
			DimStyle.Render("Insert a USB drive and press r to refresh")
		return BoxStyle.Render(content)
	}

	s := DimStyle.Render(fmt.Sprintf("Found %d USB drive(s):", len(m.drives))) + "\n\n"

	for i, drv := range m.drives {
		cursor := "  "
		nameStyle := NormalItemStyle
		detailStyle := DimStyle

		if i == m.cursor {
			cursor = SelectedStyle.Render("> ")
			nameStyle = ActiveItemStyle
			detailStyle = lipgloss.NewStyle().Foreground(ColorYellow)
		}

		s += cursor + nameStyle.Render(drv.DisplayName()) + "\n"
		s += detailStyle.Render(fmt.Sprintf("    %s  %s", drv.SizeHuman(), drv.DeviceNode)) + "\n"
		if len(drv.MountPoints) > 0 {
			s += detailStyle.Render(fmt.Sprintf("    Mounted: %s", drv.MountPoints[0])) + "\n"
		}
		if i < len(m.drives)-1 {
			s += "\n"
		}
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorOrange).
		Padding(1, 2).
		Render(s)
}
