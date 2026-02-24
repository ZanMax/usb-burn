package tui

import (
	"fmt"

	"usb_burn/internal/device"
	"usb_burn/internal/image"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type filePickerModel struct {
	images   []image.ImageFile
	cursor   int
	loading  bool
	err      string
	selected bool
	dir      string
}

func newFilePickerModel(dir string) filePickerModel {
	return filePickerModel{
		loading: true,
		dir:     dir,
	}
}

func (m filePickerModel) Init() tea.Cmd {
	return m.scanImages()
}

func (m filePickerModel) scanImages() tea.Cmd {
	dir := m.dir
	return func() tea.Msg {
		images, err := image.ScanDirectory(dir)
		if err != nil {
			return imagesScannedMsg{Err: err}
		}
		var infos []ImageInfo
		for _, img := range images {
			infos = append(infos, ImageInfo{
				Name: img.Name,
				Path: img.Path,
				Size: img.Size,
			})
		}
		return imagesScannedMsg{Images: infos}
	}
}

func (m filePickerModel) Update(msg tea.Msg) (filePickerModel, tea.Cmd) {
	switch msg := msg.(type) {
	case imagesScannedMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err.Error()
			return m, nil
		}
		if len(msg.Images) == 0 {
			m.err = "No image files found (.iso, .img, .dmg, .raw)"
		}
		m.images = make([]image.ImageFile, len(msg.Images))
		for i, info := range msg.Images {
			m.images[i] = image.ImageFile{
				Name: info.Name,
				Path: info.Path,
				Size: info.Size,
			}
		}
		return m, nil

	case tea.KeyMsg:
		if m.loading {
			return m, nil
		}
		switch {
		case key.Matches(msg, Keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, Keys.Down):
			if m.cursor < len(m.images)-1 {
				m.cursor++
			}
		case key.Matches(msg, Keys.Select):
			if len(m.images) > 0 {
				m.selected = true
			}
		}
	}
	return m, nil
}

func (m filePickerModel) View() string {
	if m.loading {
		return BoxStyle.Render("Scanning for image files...")
	}
	if m.err != "" {
		return BoxStyle.Render(ErrorStyle.Render(m.err) + "\n\n" + DimStyle.Render("Press esc to go back and choose another directory"))
	}

	s := DimStyle.Render(fmt.Sprintf("Found %d image file(s) in %s:", len(m.images), m.dir)) + "\n\n"

	for i, img := range m.images {
		cursor := "  "
		style := NormalItemStyle
		sizeStyle := DimStyle

		if i == m.cursor {
			cursor = SelectedStyle.Render("> ")
			style = ActiveItemStyle
			sizeStyle = lipgloss.NewStyle().Foreground(ColorYellow)
		}

		size := device.FormatBytes(img.Size)
		s += cursor + style.Render(img.Name) + sizeStyle.Render(fmt.Sprintf("  (%s)", size)) + "\n"
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorOrange).
		Padding(1, 2).
		Render(s)
}
