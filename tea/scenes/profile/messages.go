package profile

import (
	tea "github.com/charmbracelet/bubbletea"
)

type updateProfileList struct{}

func updateProfileListCmd() tea.Msg {
	return updateProfileList{}
}

type updateProfileNames struct{}

func updateProfileNamesCmd() tea.Msg {
	return updateProfileNames{}
}
