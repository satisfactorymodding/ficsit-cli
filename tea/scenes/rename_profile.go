package scenes

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/satisfactorymodding/ficsit-cli/cli"
	"github.com/satisfactorymodding/ficsit-cli/tea/components"
	"github.com/satisfactorymodding/ficsit-cli/tea/utils"
)

var _ tea.Model = (*renameProfile)(nil)

type renameProfile struct {
	root    components.RootModel
	parent  tea.Model
	input   textinput.Model
	title   string
	oldName string
}

func NewRenameProfile(root components.RootModel, parent tea.Model, profileData *cli.Profile) tea.Model {
	model := renameProfile{
		root:    root,
		parent:  parent,
		input:   textinput.NewModel(),
		title:   utils.NonListTitleStyle.Render(fmt.Sprintf("Rename Profile: %s", profileData.Name)),
		oldName: profileData.Name,
	}

	model.input.SetValue(profileData.Name)
	model.input.Focus()
	model.input.Width = root.Size().Width

	return model
}

func (m renameProfile) Init() tea.Cmd {
	return nil
}

func (m renameProfile) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case KeyControlC:
			return m, tea.Quit
		case KeyEscape:
			return m.parent, nil
		case KeyEnter:
			if err := m.root.GetGlobal().Profiles.RenameProfile(m.oldName, m.input.Value()); err != nil {
				panic(err) // TODO Handle Error
			}

			return m.parent, updateProfileNamesCmd
		default:
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			return m, cmd
		}
	case tea.WindowSizeMsg:
		m.root.SetSize(msg)
	}

	return m, nil
}

func (m renameProfile) View() string {
	inputView := lipgloss.NewStyle().Padding(1, 2).Render(m.input.View())
	return lipgloss.JoinVertical(lipgloss.Left, m.root.View(), m.title, inputView)
}
