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
	currentProfile      *cli.Profile
	currentInstallation *cli.Installation
	global              *cli.GlobalContext
	apiClient           graphql.Client
	currentSize         tea.WindowSizeMsg
	headerComponent     tea.Model
}

func newModel(global *cli.GlobalContext) *rootModel {
	m := &rootModel{
		global:              global,
		currentProfile:      global.Profiles.GetProfile(global.Profiles.SelectedProfile),
		currentInstallation: global.Installations.GetInstallation(global.Installations.SelectedInstallation),
		apiClient:           ficsit.InitAPI(),
		currentSize: tea.WindowSizeMsg{
			Width:  20,
			Height: 14,
		},
	}

	m.headerComponent = components.NewHeaderComponent(m)

	return m
}

func (m *rootModel) GetCurrentProfile() *cli.Profile {
	return m.currentProfile
}

func (m *rootModel) SetCurrentProfile(profile *cli.Profile) error {
	m.currentProfile = profile
	m.global.Profiles.SelectedProfile = profile.Name
	return m.global.Save()
}

func (m *rootModel) GetCurrentInstallation() *cli.Installation {
	return m.currentInstallation
}

func (m *rootModel) SetCurrentInstallation(installation *cli.Installation) error {
	m.currentInstallation = installation
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
	if err := tea.NewProgram(scenes.NewMainMenu(newModel(global))).Start(); err != nil {
		return errors.Wrap(err, "internal tea error")
	}
	return nil
}
