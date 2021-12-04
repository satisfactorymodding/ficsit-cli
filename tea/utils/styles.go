package utils

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

var (
	ListStyles        list.Styles
	LabelStyle        = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("202"))
	TitleStyle        = list.DefaultStyles().Title.Background(lipgloss.Color("22"))
	NonListTitleStyle = TitleStyle.Copy().MarginLeft(2).Background(lipgloss.Color("22"))
)

func init() {
	ListStyles = list.DefaultStyles()
	ListStyles.Title = TitleStyle
	ListStyles.HelpStyle = list.DefaultStyles().HelpStyle.PaddingLeft(2).PaddingBottom(1)
	ListStyles.PaginationStyle = list.DefaultStyles().PaginationStyle.PaddingLeft(2)
}
