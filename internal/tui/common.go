package tui

import (
	"usb_burn/internal/device"

	tea "github.com/charmbracelet/bubbletea"
)

// Screen identifies which screen is currently active.
type Screen int

const (
	ScreenMenu Screen = iota
	ScreenDirPicker
	ScreenFilePicker
	ScreenDrivePicker
	ScreenFormatPicker
	ScreenConfirm
	ScreenProgress
	ScreenDone
)

// Mode identifies which flow the user is in.
type Mode int

const (
	ModeWrite Mode = iota
	ModeFormat
)

// AppState holds the accumulated wizard selections.
type AppState struct {
	Mode           Mode
	Dir            string
	SelectedImage  string
	ImageSize      int64
	SelectedDrive  *device.Drive
	SelectedFormat string
	FormatLabel    string
	Error          error
	BytesWritten   int64
	Duration       float64
}

// Messages used for screen transitions and async updates.

type switchScreenMsg struct {
	screen Screen
}

type progressUpdateMsg struct {
	Phase   string
	Percent float64
	Bytes   int64
	Total   int64
	Speed   float64 // bytes per second
}

type operationDoneMsg struct {
	Err          error
	BytesWritten int64
	Duration     float64
}

type drivesDetectedMsg struct {
	Drives []device.Drive
	Err    error
}

type imagesScannedMsg struct {
	Images []ImageInfo
	Err    error
}

// ImageInfo holds metadata about a discovered image file.
type ImageInfo struct {
	Name string
	Path string
	Size int64
}

// SwitchScreen returns a command to transition to a new screen.
func SwitchScreen(s Screen) tea.Cmd {
	return func() tea.Msg {
		return switchScreenMsg{screen: s}
	}
}
