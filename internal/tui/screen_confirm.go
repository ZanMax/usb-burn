package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type confirmModel struct {
	confirmed bool
	cancelled bool
}

func newConfirmModel() confirmModel {
	return confirmModel{}
}

func (m confirmModel) Init() tea.Cmd {
	return nil
}

func (m confirmModel) Update(msg tea.Msg) (confirmModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, Keys.Yes):
			m.confirmed = true
		case key.Matches(msg, Keys.No), key.Matches(msg, Keys.Back):
			m.cancelled = true
		}
	}
	return m, nil
}

func (m confirmModel) ViewWrite(state *AppState) string {
	imageName := state.SelectedImage
	driveName := state.SelectedDrive.DisplayName()
	driveNode := state.SelectedDrive.DeviceNode

	s := WarningStyle.Render("WARNING: ALL DATA WILL BE DESTROYED") + "\n\n"
	s += fmt.Sprintf("Image:  %s\n", SelectedStyle.Render(imageName))
	s += fmt.Sprintf("Drive:  %s (%s)\n", SelectedStyle.Render(driveName), driveNode)
	s += fmt.Sprintf("Size:   %s\n", state.SelectedDrive.SizeHuman())
	s += "\n"
	s += "This will overwrite ALL data on the target drive.\n"
	s += "This action cannot be undone.\n\n"
	s += SelectedStyle.Render("Press y to confirm, n or esc to cancel")

	return DangerBoxStyle.Render(s)
}

func (m confirmModel) ViewFormat(state *AppState) string {
	driveName := state.SelectedDrive.DisplayName()
	driveNode := state.SelectedDrive.DeviceNode
	fsName := state.FormatLabel

	s := WarningStyle.Render("WARNING: ALL DATA WILL BE DESTROYED") + "\n\n"
	s += fmt.Sprintf("Drive:   %s (%s)\n", SelectedStyle.Render(driveName), driveNode)
	s += fmt.Sprintf("Size:    %s\n", state.SelectedDrive.SizeHuman())
	s += fmt.Sprintf("Format:  %s\n", SelectedStyle.Render(fsName))
	s += "\n"
	s += "This will erase ALL data and format the drive.\n"
	s += "This action cannot be undone.\n\n"
	s += SelectedStyle.Render("Press y to confirm, n or esc to cancel")

	return DangerBoxStyle.Render(s)
}
