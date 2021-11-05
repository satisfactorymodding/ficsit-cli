package tea

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/satisfactorymodding/ficsit-cli/cli"
	"github.com/satisfactorymodding/ficsit-cli/tea/scenes"
	"os"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type item string

func (i item) FilterValue() string { return string(i) }

type rootModel struct {
	currentModel        tea.Model
	currentProfile      *cli.Profile
	currentInstallation *cli.Installation
}

func (m *rootModel) ChangeScene(model tea.Model) {
	m.currentModel = model
}

func (m *rootModel) GetCurrentProfile() *cli.Profile {
	return m.currentProfile
}

func (m *rootModel) SetCurrentProfile(profile *cli.Profile) {
	m.currentProfile = profile
}

func (m *rootModel) GetCurrentInstallation() *cli.Installation {
	return m.currentInstallation
}

func (m *rootModel) SetCurrentInstallation(installation *cli.Installation) {
	m.currentInstallation = installation
}

func newModel() rootModel {
	m := rootModel{}
	m.currentModel = scenes.NewMainMenu(&m)
	return m
}

func (m rootModel) Init() tea.Cmd {
	return m.currentModel.Init()
}

func (m rootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.currentModel, cmd = m.currentModel.Update(msg)
	return m, cmd
}

func (m rootModel) View() string {

	style := lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("220"))

	out := style.Render("Installation:") + " " + "// TODO" + "\n"
	out += style.Render("Profile:") + " " + "// TODO" + "\n"
	out += "\n"

	return out + m.currentModel.View()
}

func RunTea() {
	if err := tea.NewProgram(newModel()).Start(); err != nil {
		fmt.Printf("Could not start program :(\n%v\n", err)
		os.Exit(1)
	}
}
