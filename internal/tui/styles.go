package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors — orange/red "burn" palette
	ColorOrange    = lipgloss.Color("#FF6B35")
	ColorRed       = lipgloss.Color("#D32F2F")
	ColorYellow    = lipgloss.Color("#FFA726")
	ColorGreen     = lipgloss.Color("#66BB6A")
	ColorDim       = lipgloss.Color("#888888")
	ColorWhite     = lipgloss.Color("#FFFFFF")
	ColorDarkGray  = lipgloss.Color("#333333")
	ColorLightGray = lipgloss.Color("#CCCCCC")

	// Title
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorOrange).
			MarginBottom(1)

	// Breadcrumb
	BreadcrumbStyle = lipgloss.NewStyle().
			Foreground(ColorDim).
			MarginBottom(1)

	// Box style for content areas
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorOrange).
			Padding(1, 2)

	// Danger box for confirmation warnings
	DangerBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorRed).
			Padding(1, 2)

	// Success box
	SuccessBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorGreen).
			Padding(1, 2)

	// Selected item highlight
	SelectedStyle = lipgloss.NewStyle().
			Foreground(ColorOrange).
			Bold(true)

	// Dimmed text
	DimStyle = lipgloss.NewStyle().
			Foreground(ColorDim)

	// Error text
	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorRed).
			Bold(true)

	// Help footer
	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorDim).
			MarginTop(1)

	// Active list item
	ActiveItemStyle = lipgloss.NewStyle().
			Foreground(ColorOrange).
			Bold(true).
			PaddingLeft(2)

	// Normal list item
	NormalItemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	// Description text
	DescStyle = lipgloss.NewStyle().
			Foreground(ColorDim).
			PaddingLeft(4)

	// Header bar
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorWhite).
			Background(ColorOrange).
			Padding(0, 2).
			MarginBottom(1)

	// Warning text
	WarningStyle = lipgloss.NewStyle().
			Foreground(ColorYellow).
			Bold(true)
)
