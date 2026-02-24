package tui

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// RootModel is the top-level model managing all screens.
type RootModel struct {
	state  AppState
	screen Screen

	// Sub-models for each screen
	menu         menuModel
	dirPicker    dirPickerModel
	filePicker   filePickerModel
	drivePicker  drivePickerModel
	formatPicker formatPickerModel
	confirm      confirmModel
	progress     progressModel
	done         doneModel

	width  int
	height int
}

// NewRootModel creates the initial root model.
func NewRootModel(defaultDir string) RootModel {
	return RootModel{
		state: AppState{
			Dir: defaultDir,
		},
		screen: ScreenMenu,
		menu:   newMenuModel(),
	}
}

func (m RootModel) Init() tea.Cmd {
	return m.menu.Init()
}

func (m RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// Global quit
		if key.Matches(msg, Keys.Quit) && m.screen != ScreenProgress && m.screen != ScreenDirPicker {
			return m, tea.Quit
		}
		// Global back
		if key.Matches(msg, Keys.Back) && m.screen != ScreenProgress && m.screen != ScreenDone && m.screen != ScreenDirPicker {
			return m.goBack()
		}
	}

	// Delegate to current screen
	switch m.screen {
	case ScreenMenu:
		return m.updateMenu(msg)
	case ScreenDirPicker:
		return m.updateDirPicker(msg)
	case ScreenFilePicker:
		return m.updateFilePicker(msg)
	case ScreenDrivePicker:
		return m.updateDrivePicker(msg)
	case ScreenFormatPicker:
		return m.updateFormatPicker(msg)
	case ScreenConfirm:
		return m.updateConfirm(msg)
	case ScreenProgress:
		return m.updateProgress(msg)
	case ScreenDone:
		return m.updateDone(msg)
	}

	return m, nil
}

// Screen update methods

func (m RootModel) updateMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.menu, cmd = m.menu.Update(msg)

	if m.menu.selected {
		m.menu.selected = false
		item := m.menu.items[m.menu.cursor]
		m.state.Mode = item.mode

		if m.state.Mode == ModeWrite {
			m.screen = ScreenDirPicker
			m.dirPicker = newDirPickerModel(m.state.Dir)
			return m, m.dirPicker.Init()
		}
		// Format flow goes straight to drive picker
		m.screen = ScreenDrivePicker
		m.drivePicker = newDrivePickerModel()
		return m, m.drivePicker.Init()
	}

	return m, cmd
}

func (m RootModel) updateDirPicker(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle quit separately for text input screen
	if msg, ok := msg.(tea.KeyMsg); ok {
		if key.Matches(msg, Keys.Quit) && msg.String() != "esc" {
			return m, tea.Quit
		}
		if key.Matches(msg, Keys.Back) {
			m.screen = ScreenMenu
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.dirPicker, cmd = m.dirPicker.Update(msg)

	if m.dirPicker.confirmed {
		m.dirPicker.confirmed = false
		m.state.Dir = m.dirPicker.textInput.Value()
		m.screen = ScreenFilePicker
		m.filePicker = newFilePickerModel(m.state.Dir)
		return m, m.filePicker.Init()
	}

	return m, cmd
}

func (m RootModel) updateFilePicker(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.filePicker, cmd = m.filePicker.Update(msg)

	if m.filePicker.selected {
		m.filePicker.selected = false
		img := m.filePicker.images[m.filePicker.cursor]
		m.state.SelectedImage = img.Path
		m.state.ImageSize = img.Size

		m.screen = ScreenDrivePicker
		m.drivePicker = newDrivePickerModel()
		return m, m.drivePicker.Init()
	}

	return m, cmd
}

func (m RootModel) updateDrivePicker(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.drivePicker, cmd = m.drivePicker.Update(msg)

	if m.drivePicker.selected {
		m.drivePicker.selected = false
		drv := m.drivePicker.drives[m.drivePicker.cursor]
		m.state.SelectedDrive = &drv

		if m.state.Mode == ModeFormat {
			m.screen = ScreenFormatPicker
			m.formatPicker = newFormatPickerModel()
			return m, m.formatPicker.Init()
		}

		m.screen = ScreenConfirm
		m.confirm = newConfirmModel()
		return m, m.confirm.Init()
	}

	return m, cmd
}

func (m RootModel) updateFormatPicker(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.formatPicker, cmd = m.formatPicker.Update(msg)

	if m.formatPicker.selected {
		m.formatPicker.selected = false
		fs := m.formatPicker.formats[m.formatPicker.cursor]
		m.state.SelectedFormat = fs.ID
		m.state.FormatLabel = fs.Name

		m.screen = ScreenConfirm
		m.confirm = newConfirmModel()
		return m, m.confirm.Init()
	}

	return m, cmd
}

func (m RootModel) updateConfirm(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.confirm, cmd = m.confirm.Update(msg)

	if m.confirm.confirmed {
		m.confirm.confirmed = false
		m.screen = ScreenProgress
		m.progress = newProgressModel(&m.state)

		if m.state.Mode == ModeWrite {
			var pCmd tea.Cmd
			m.progress, pCmd = m.progress.startWriteOperation(&m.state)
			return m, pCmd
		}
		var pCmd tea.Cmd
		m.progress, pCmd = m.progress.startFormatOperation(&m.state)
		return m, pCmd
	}

	if m.confirm.cancelled {
		m.confirm.cancelled = false
		return m.goBack()
	}

	return m, cmd
}

func (m RootModel) updateProgress(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.progress, cmd = m.progress.Update(msg)

	if m.progress.done {
		m.state.Error = m.progress.err
		m.state.BytesWritten = m.progress.bytes
		m.state.Duration = time.Since(m.progress.startTime).Seconds()

		m.screen = ScreenDone
		m.done = newDoneModel()
		return m, m.done.Init()
	}

	return m, cmd
}

func (m RootModel) updateDone(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.done, cmd = m.done.Update(msg)

	if m.done.restarted {
		m.done.restarted = false
		m.state = AppState{Dir: m.state.Dir}
		m.screen = ScreenMenu
		m.menu = newMenuModel()
		return m, m.menu.Init()
	}

	return m, cmd
}

// goBack navigates to the previous screen based on the current flow.
func (m RootModel) goBack() (tea.Model, tea.Cmd) {
	switch m.screen {
	case ScreenDirPicker:
		m.screen = ScreenMenu
		return m, nil
	case ScreenFilePicker:
		m.screen = ScreenDirPicker
		m.dirPicker = newDirPickerModel(m.state.Dir)
		return m, m.dirPicker.Init()
	case ScreenDrivePicker:
		if m.state.Mode == ModeWrite {
			m.screen = ScreenFilePicker
			m.filePicker = newFilePickerModel(m.state.Dir)
			return m, m.filePicker.Init()
		}
		m.screen = ScreenMenu
		return m, nil
	case ScreenFormatPicker:
		m.screen = ScreenDrivePicker
		m.drivePicker = newDrivePickerModel()
		return m, m.drivePicker.Init()
	case ScreenConfirm:
		if m.state.Mode == ModeWrite {
			m.screen = ScreenDrivePicker
			m.drivePicker = newDrivePickerModel()
			return m, m.drivePicker.Init()
		}
		m.screen = ScreenFormatPicker
		m.formatPicker = newFormatPickerModel()
		return m, m.formatPicker.Init()
	}
	return m, nil
}

func (m RootModel) View() string {
	header := m.renderHeader()
	body := m.renderBody()
	footer := m.renderFooter()

	return lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
}

func (m RootModel) renderHeader() string {
	title := HeaderStyle.Render(" USB BURN ")
	breadcrumb := m.renderBreadcrumb()

	return lipgloss.JoinVertical(lipgloss.Left, title, breadcrumb)
}

func (m RootModel) renderBreadcrumb() string {
	parts := []string{}

	switch m.screen {
	case ScreenMenu:
		parts = append(parts, "Main Menu")
	case ScreenDirPicker:
		parts = append(parts, "Write Image", "Select Directory")
	case ScreenFilePicker:
		parts = append(parts, "Write Image", "Select Image")
	case ScreenDrivePicker:
		if m.state.Mode == ModeWrite {
			parts = append(parts, "Write Image", "Select Drive")
		} else {
			parts = append(parts, "Format Drive", "Select Drive")
		}
	case ScreenFormatPicker:
		parts = append(parts, "Format Drive", "Select Format")
	case ScreenConfirm:
		if m.state.Mode == ModeWrite {
			parts = append(parts, "Write Image", "Confirm")
		} else {
			parts = append(parts, "Format Drive", "Confirm")
		}
	case ScreenProgress:
		if m.state.Mode == ModeWrite {
			parts = append(parts, "Write Image", "Writing...")
		} else {
			parts = append(parts, "Format Drive", "Formatting...")
		}
	case ScreenDone:
		if m.state.Mode == ModeWrite {
			parts = append(parts, "Write Image", "Done")
		} else {
			parts = append(parts, "Format Drive", "Done")
		}
	}

	crumb := ""
	for i, p := range parts {
		if i > 0 {
			crumb += DimStyle.Render(" > ")
		}
		if i == len(parts)-1 {
			crumb += SelectedStyle.Render(p)
		} else {
			crumb += DimStyle.Render(p)
		}
	}

	return BreadcrumbStyle.Render(crumb)
}

func (m RootModel) renderBody() string {
	switch m.screen {
	case ScreenMenu:
		return m.menu.View()
	case ScreenDirPicker:
		return m.dirPicker.View()
	case ScreenFilePicker:
		return m.filePicker.View()
	case ScreenDrivePicker:
		return m.drivePicker.View()
	case ScreenFormatPicker:
		return m.formatPicker.View()
	case ScreenConfirm:
		if m.state.Mode == ModeWrite {
			return m.confirm.ViewWrite(&m.state)
		}
		return m.confirm.ViewFormat(&m.state)
	case ScreenProgress:
		return m.progress.View()
	case ScreenDone:
		if m.state.Error != nil {
			return m.done.ViewError(m.state.Error)
		}
		return m.done.ViewSuccess(&m.state)
	}
	return ""
}

func (m RootModel) renderFooter() string {
	help := ""
	switch m.screen {
	case ScreenMenu:
		help = "↑/↓ navigate  enter select  q quit"
	case ScreenDirPicker:
		help = "enter confirm  esc back  ctrl+c quit"
	case ScreenFilePicker:
		help = "↑/↓ navigate  enter select  esc back  q quit"
	case ScreenDrivePicker:
		help = "↑/↓ navigate  enter select  r refresh  esc back  q quit"
	case ScreenFormatPicker:
		help = "↑/↓ navigate  enter select  esc back  q quit"
	case ScreenConfirm:
		help = "y confirm  n/esc cancel  q quit"
	case ScreenProgress:
		help = "please wait..."
	case ScreenDone:
		help = "r restart  q/enter quit"
	}

	// Show current selection info
	info := m.renderSelectionInfo()
	if info != "" {
		return HelpStyle.Render(info) + "\n" + HelpStyle.Render(help)
	}

	return HelpStyle.Render(help)
}

func (m RootModel) renderSelectionInfo() string {
	parts := []string{}

	if m.state.Mode == ModeWrite && m.state.SelectedImage != "" && m.screen != ScreenConfirm && m.screen != ScreenProgress && m.screen != ScreenDone {
		parts = append(parts, fmt.Sprintf("Image: %s", filepath.Base(m.state.SelectedImage)))
	}
	if m.state.SelectedDrive != nil && m.screen != ScreenConfirm && m.screen != ScreenProgress && m.screen != ScreenDone {
		parts = append(parts, fmt.Sprintf("Drive: %s", m.state.SelectedDrive.DisplayName()))
	}

	if len(parts) == 0 {
		return ""
	}

	result := ""
	for _, p := range parts {
		if result != "" {
			result += "  |  "
		}
		result += p
	}
	return result
}
