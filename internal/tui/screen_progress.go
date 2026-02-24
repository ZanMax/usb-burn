package tui

import (
	"fmt"
	"time"

	"usb_burn/internal/device"
	"usb_burn/internal/formatter"
	"usb_burn/internal/writer"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// channelSubscriptionMsg is used to poll a channel for updates.
type channelSubscriptionMsg struct{}

type progressModel struct {
	progress  progress.Model
	spinner   spinner.Model
	phase     string
	percent   float64
	bytes     int64
	total     int64
	speed     float64
	done      bool
	err       error
	startTime time.Time

	// Channel-based communication
	writeCh  chan writer.ProgressUpdate
	formatCh chan formatter.ProgressUpdate
	mode     Mode
}

func newProgressModel(state *AppState) progressModel {
	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(50),
	)
	p.FullColor = string(ColorOrange)
	p.EmptyColor = string(ColorDarkGray)

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(ColorOrange)

	m := progressModel{
		progress:  p,
		spinner:   s,
		phase:     "Starting...",
		startTime: time.Now(),
		mode:      state.Mode,
	}

	return m
}

func (m progressModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.pollChannel())
}

func (m progressModel) startWriteOperation(state *AppState) (progressModel, tea.Cmd) {
	m.writeCh = make(chan writer.ProgressUpdate, 10)
	go writer.WriteImage(state.SelectedImage, state.SelectedDrive, m.writeCh)
	return m, tea.Batch(m.spinner.Tick, m.pollChannel())
}

func (m progressModel) startFormatOperation(state *AppState) (progressModel, tea.Cmd) {
	m.formatCh = make(chan formatter.ProgressUpdate, 10)
	go formatter.FormatDrive(state.SelectedDrive, state.SelectedFormat, "USB_BURN", m.formatCh)
	return m, tea.Batch(m.spinner.Tick, m.pollChannel())
}

func (m progressModel) pollChannel() tea.Cmd {
	return func() tea.Msg {
		// Small delay to not busy-loop
		time.Sleep(50 * time.Millisecond)
		return channelSubscriptionMsg{}
	}
}

func (m progressModel) Update(msg tea.Msg) (progressModel, tea.Cmd) {
	switch msg := msg.(type) {
	case channelSubscriptionMsg:
		if m.done {
			return m, nil
		}

		if m.mode == ModeWrite && m.writeCh != nil {
			return m.drainWriteChannel()
		}
		if m.mode == ModeFormat && m.formatCh != nil {
			return m.drainFormatChannel()
		}
		return m, m.pollChannel()

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case progress.FrameMsg:
		pm, cmd := m.progress.Update(msg)
		m.progress = pm.(progress.Model)
		return m, cmd
	}
	return m, nil
}

func (m progressModel) drainWriteChannel() (progressModel, tea.Cmd) {
	var lastUpdate *writer.ProgressUpdate
	// Drain all pending messages from the channel
	for {
		select {
		case update, ok := <-m.writeCh:
			if !ok {
				// Channel closed
				if !m.done {
					m.done = true
				}
				return m, nil
			}
			lastUpdate = &update
			if update.Done {
				m.done = true
				m.err = update.Err
				m.phase = update.Phase
				m.percent = update.Percent
				m.bytes = update.Bytes
				m.total = update.Total
				m.speed = update.Speed
				return m, nil
			}
		default:
			// No more messages
			if lastUpdate != nil {
				m.phase = lastUpdate.Phase
				m.percent = lastUpdate.Percent
				m.bytes = lastUpdate.Bytes
				m.total = lastUpdate.Total
				m.speed = lastUpdate.Speed
			}
			return m, tea.Batch(m.spinner.Tick, m.pollChannel())
		}
	}
}

func (m progressModel) drainFormatChannel() (progressModel, tea.Cmd) {
	var lastUpdate *formatter.ProgressUpdate
	for {
		select {
		case update, ok := <-m.formatCh:
			if !ok {
				if !m.done {
					m.done = true
				}
				return m, nil
			}
			lastUpdate = &update
			if update.Done {
				m.done = true
				m.err = update.Err
				m.phase = update.Phase
				return m, nil
			}
		default:
			if lastUpdate != nil {
				m.phase = lastUpdate.Phase
			}
			return m, tea.Batch(m.spinner.Tick, m.pollChannel())
		}
	}
}

func (m progressModel) View() string {
	s := ""

	if m.mode == ModeWrite {
		s += m.spinner.View() + " " + SelectedStyle.Render(m.phase) + "\n\n"
		s += m.progress.ViewAs(m.percent) + "\n\n"

		if m.total > 0 {
			s += fmt.Sprintf("  %s / %s",
				device.FormatBytes(m.bytes),
				device.FormatBytes(m.total))
			if m.speed > 0 {
				s += fmt.Sprintf("  |  %s/s", device.FormatBytes(int64(m.speed)))
			}
			if m.speed > 0 && m.bytes < m.total {
				remaining := float64(m.total-m.bytes) / m.speed
				s += fmt.Sprintf("  |  ETA: %s", formatDuration(remaining))
			}
		}
	} else {
		// Format mode - simpler progress (no byte-level tracking)
		s += m.spinner.View() + " " + SelectedStyle.Render(m.phase) + "\n\n"
		elapsed := time.Since(m.startTime).Seconds()
		s += DimStyle.Render(fmt.Sprintf("  Elapsed: %s", formatDuration(elapsed)))
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorOrange).
		Padding(1, 2).
		Render(s)
}

func formatDuration(seconds float64) string {
	if seconds < 60 {
		return fmt.Sprintf("%.0fs", seconds)
	}
	m := int(seconds) / 60
	s := int(seconds) % 60
	return fmt.Sprintf("%dm %ds", m, s)
}
