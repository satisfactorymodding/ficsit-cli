package utils

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

var (
	ListStyles        = list.DefaultStyles()
	LabelStyle        = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("202"))
	TitleStyle        = list.DefaultStyles().Title.Background(lipgloss.Color("#b34100"))
	NonListTitleStyle = TitleStyle.Copy().MarginLeft(2).Background(lipgloss.Color("#b34100"))
)

var (
	LogoForegroundStyles = []lipgloss.Style{
		lipgloss.NewStyle().Foreground(lipgloss.Color("#ff5f00")).Background(lipgloss.Color("#ff5f00")),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#e65400")).Background(lipgloss.Color("#e65400")),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#cc4b00")).Background(lipgloss.Color("#cc4b00")),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#b34100")).Background(lipgloss.Color("#b34100")),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#993800")).Background(lipgloss.Color("#993800")),
		lipgloss.NewStyle(),
	}
	LogoBackgroundStyles = []lipgloss.Style{
		lipgloss.NewStyle().Foreground(lipgloss.Color("255")),
		lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
		lipgloss.NewStyle().Foreground(lipgloss.Color("249")),
		lipgloss.NewStyle().Foreground(lipgloss.Color("246")),
		lipgloss.NewStyle().Foreground(lipgloss.Color("243")),
		lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
	}
)

func init() {
	ListStyles.Title = TitleStyle
	ListStyles.HelpStyle = list.DefaultStyles().HelpStyle.PaddingLeft(2).PaddingBottom(1)
	ListStyles.PaginationStyle = list.DefaultStyles().PaginationStyle.PaddingLeft(2)
}
