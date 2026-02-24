package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"usb_burn/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func main() {
	// Parse flags
	defaultDir := filepath.Join(os.Getenv("HOME"), "Downloads")
	dir := flag.String("dir", defaultDir, "Directory to scan for image files")
	flag.Parse()

	// Expand ~
	if len(*dir) >= 2 && (*dir)[:2] == "~/" {
		*dir = filepath.Join(os.Getenv("HOME"), (*dir)[2:])
	} else if *dir == "~" {
		*dir = os.Getenv("HOME")
	}

	// Check for root privileges
	if os.Geteuid() != 0 {
		errStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF6B35")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#D32F2F")).
			Padding(1, 2)

		fmt.Println(errStyle.Render(
			"USB Burn requires root privileges to write to devices.\n\n" +
				"Please re-run with sudo:\n" +
				"  sudo " + os.Args[0]))
		os.Exit(1)
	}

	// Create and run TUI
	model := tui.NewRootModel(*dir)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
