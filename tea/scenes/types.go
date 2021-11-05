package scenes

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/satisfactorymodding/ficsit-cli/cli"
)

type RootModel interface {
	ChangeScene(model tea.Model)

	GetCurrentProfile() *cli.Profile
	SetCurrentProfile(profile *cli.Profile)

	GetCurrentInstallation() *cli.Installation
	SetCurrentInstallation(installation *cli.Installation)
}
