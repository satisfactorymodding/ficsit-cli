package scenes

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

type updateInstallationList struct{}

func updateInstallationListCmd() tea.Msg {
	return updateInstallationList{}
}

type updateInstallationNames struct{}

func updateInstallationNamesCmd() tea.Msg {
	return updateInstallationNames{}
}
