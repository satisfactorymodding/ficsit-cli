package scenes

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/satisfactorymodding/ficsit-cli/tea/components"
	"github.com/satisfactorymodding/ficsit-cli/tea/utils"
)

var _ tea.Model = (*newInstallation)(nil)

type newInstallation struct {
	root   components.RootModel
	parent tea.Model
	input  textinput.Model
	path   string
}

func NewNewInstallation(root components.RootModel, parent tea.Model) tea.Model {
	model := newInstallation{
		root:   root,
		parent: parent,
		input:  textinput.NewModel(),
		path:  utils.NonListTitleStyle.Render("/path/to/installation"),
	}

	model.input.Focus()
	model.input.Width = root.Size().Width

	return model
}

func (m newInstallation) Init() tea.Cmd {
	return nil
}

func (m newInstallation) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case KeyControlC:
			return m, tea.Quit
		case KeyEscape:
			return m.parent, nil
		case KeyEnter:
			// Use the currently selected profile for the new installation
			var selectedProfile = m.root.GetGlobal().Profiles.SelectedProfile
			if _, err := m.root.GetGlobal().Installations.AddInstallation(m.root.GetGlobal(), m.input.Value(), selectedProfile); err != nil {
				panic(err) // TODO Handle Error
			}

			return m.parent, updateInstallationListCmd
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

func (m newInstallation) View() string {
	inputView := lipgloss.NewStyle().Padding(1, 2).Render(m.input.View())
	return lipgloss.JoinVertical(lipgloss.Left, m.root.View(), m.path, inputView)
}
