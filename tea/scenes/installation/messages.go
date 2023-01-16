package installation

import tea "github.com/charmbracelet/bubbletea"

type updateInstallationList struct{}

func updateInstallationListCmd() tea.Msg {
	return updateInstallationList{}
}

type updateInstallationNames struct{}

func updateInstallationNamesCmd() tea.Msg {
	return updateInstallationNames{}
}
