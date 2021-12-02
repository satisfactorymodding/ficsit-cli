package utils

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type TickMsg struct{}

func Ticker() tea.Cmd {
	return tea.Tick(time.Millisecond*50, func(time.Time) tea.Msg {
		return TickMsg{}
	})
}
