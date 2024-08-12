package components

import (
	"github.com/Khan/genqlient/graphql"
	tea "github.com/charmbracelet/bubbletea"
	resolver "github.com/satisfactorymodding/ficsit-resolver"

	"github.com/satisfactorymodding/ficsit-cli/cli"
	"github.com/satisfactorymodding/ficsit-cli/cli/provider"
)

type RootModel interface {
	GetGlobal() *cli.GlobalContext

	GetCurrentProfile() *cli.Profile
	SetCurrentProfile(profile *cli.Profile) error

	GetCurrentInstallation() *cli.Installation
	SetCurrentInstallation(installation *cli.Installation) error

	GetAPIClient() graphql.Client
	GetProvider() provider.Provider
	GetResolver() resolver.DependencyResolver

	Size() tea.WindowSizeMsg
	SetSize(size tea.WindowSizeMsg)

	View() string
	Height() int
}
