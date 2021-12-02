package utils

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var _ ListItem = (*SimpleItem)(nil)
var _ list.Item = (*SimpleItem)(nil)

type SimpleItem struct {
	Title    string
	Activate func(msg tea.Msg, currentModel tea.Model) (tea.Model, tea.Cmd)
}

func (n SimpleItem) FilterValue() string {
	return n.Title
}

func (n SimpleItem) GetTitle() string {
	return n.Title
}

type ListItem interface {
	GetTitle() string
}

type ItemDelegate struct{}

func (d ItemDelegate) Height() int                               { return 1 }
func (d ItemDelegate) Spacing() int                              { return 0 }
func (d ItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d ItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(ListItem)
	if !ok {
		return
	}

	style := lipgloss.NewStyle().PaddingLeft(2)

	str := style.Render("o " + i.GetTitle())
	if index == m.Index() {
		str = style.Foreground(lipgloss.Color("202")).Render("• " + i.GetTitle())
	}

	fmt.Fprint(w, str)
}
