package tea

import (
	"github.com/Khan/genqlient/graphql"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pkg/errors"
	"github.com/satisfactorymodding/ficsit-cli/cli"
	"github.com/satisfactorymodding/ficsit-cli/ficsit"
	"github.com/satisfactorymodding/ficsit-cli/tea/components"
	"github.com/satisfactorymodding/ficsit-cli/tea/scenes"
)

type rootModel struct {
	global          *cli.GlobalContext
	apiClient       graphql.Client
	currentSize     tea.WindowSizeMsg
	headerComponent tea.Model
}

func newModel(global *cli.GlobalContext) *rootModel {
	m := &rootModel{
		global:    global,
		apiClient: ficsit.InitAPI(),
		currentSize: tea.WindowSizeMsg{
			Width:  20,
			Height: 14,
		},
	}

	m.headerComponent = components.NewHeaderComponent(m)

	return m
}

func (m *rootModel) GetCurrentProfile() *cli.Profile {
	return m.global.Profiles.GetProfile(m.global.Profiles.SelectedProfile)
}

func (m *rootModel) SetCurrentProfile(profile *cli.Profile) error {
	m.global.Profiles.SelectedProfile = profile.Name
	return m.global.Save()
}

func (m *rootModel) GetCurrentInstallation() *cli.Installation {
	return m.global.Installations.GetInstallation(m.global.Installations.SelectedInstallation)
}

func (m *rootModel) SetCurrentInstallation(installation *cli.Installation) error {
	m.global.Installations.SelectedInstallation = installation.Path
	return m.global.Save()
}

func (m *rootModel) GetAPIClient() graphql.Client {
	return m.apiClient
}

func (m *rootModel) Size() tea.WindowSizeMsg {
	return m.currentSize
}

func (m *rootModel) SetSize(size tea.WindowSizeMsg) {
	m.currentSize = size
}

func (m *rootModel) View() string {
	return m.headerComponent.View()
}

func (m *rootModel) Height() int {
	return lipgloss.Height(m.View()) + 1
}

func (m *rootModel) GetGlobal() *cli.GlobalContext {
	return m.global
}

func RunTea(global *cli.GlobalContext) error {
	if err := tea.NewProgram(scenes.NewMainMenu(newModel(global)), tea.WithAltScreen(), tea.WithMouseCellMotion()).Start(); err != nil {
		return errors.Wrap(err, "internal tea error")
	}
	return nil
}
