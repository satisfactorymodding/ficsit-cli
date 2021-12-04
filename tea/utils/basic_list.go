package utils

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var _ list.DefaultItem = (*SimpleItem)(nil)

type SimpleItem struct {
	ItemTitle string
	Activate  func(msg tea.Msg, currentModel tea.Model) (tea.Model, tea.Cmd)

	// I know this is ugly but generics are coming soon and I cba
	Extra interface{}
}

func (n SimpleItem) Title() string {
	return n.ItemTitle
}

func (n SimpleItem) FilterValue() string {
	return n.ItemTitle
}

func (n SimpleItem) GetTitle() string {
	return n.ItemTitle
}

func (n SimpleItem) Description() string {
	return ""
}

func NewItemDelegate() list.ItemDelegate {
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	delegate.SetSpacing(0)

	// TODO Adaptive Colors
	// TODO Description Colors
	delegate.Styles.NormalTitle = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#1a1a1a", Dark: "#dddddd"}).
		Padding(0, 0, 0, 2)

	delegate.Styles.DimmedTitle = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#A49FA5", Dark: "#777777"}).
		Padding(0, 0, 0, 2)

	delegate.Styles.SelectedTitle = lipgloss.NewStyle().
		Border(lipgloss.ThickBorder(), false, false, false, true).
		BorderForeground(lipgloss.Color("202")).
		Foreground(lipgloss.Color("202")).
		Padding(0, 0, 0, 1)

	return delegate
}
