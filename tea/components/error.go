package components

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/satisfactorymodding/ficsit-cli/tea/utils"
)

var _ tea.Model = (*ErrorComponent)(nil)

type ErrorComponent struct {
	message    string
	labelStyle lipgloss.Style
}

func NewErrorComponent(message string, duration time.Duration) (*ErrorComponent, tea.Cmd) {
	timer := time.NewTimer(duration)

	return &ErrorComponent{
			message:    message,
			labelStyle: utils.LabelStyle,
		}, func() tea.Msg {
			<-timer.C
			return ErrorComponentTimeoutMsg{}
		}
}

func (e ErrorComponent) Init() tea.Cmd {
	return nil
}

func (e ErrorComponent) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return e, nil
}

func (e ErrorComponent) View() string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(lipgloss.Color("196")).
		Padding(0, 1).
		Margin(0, 0, 1, 2).
		Render(e.message)
}

type ErrorComponentTimeoutMsg struct{}
