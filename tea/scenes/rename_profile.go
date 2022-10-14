package scenes

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/satisfactorymodding/ficsit-cli/cli"
	"github.com/satisfactorymodding/ficsit-cli/tea/components"
	"github.com/satisfactorymodding/ficsit-cli/tea/utils"
)

var _ tea.Model = (*renameProfile)(nil)

type renameProfile struct {
	input   textinput.Model
	root    components.RootModel
	parent  tea.Model
	error   *components.ErrorComponent
	title   string
	oldName string
}

func NewRenameProfile(root components.RootModel, parent tea.Model, profileData *cli.Profile) tea.Model {
	model := renameProfile{
		root:    root,
		parent:  parent,
		input:   textinput.New(),
		title:   utils.NonListTitleStyle.Render(fmt.Sprintf("Rename Profile: %s", profileData.Name)),
		oldName: profileData.Name,
	}

	model.input.SetValue(profileData.Name)
	model.input.Focus()
	model.input.Width = root.Size().Width

	return model
}

func (m renameProfile) Init() tea.Cmd {
	return textinput.Blink
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
			if err := m.root.GetGlobal().Profiles.RenameProfile(m.root.GetGlobal(), m.oldName, m.input.Value()); err != nil {
				errorComponent, cmd := components.NewErrorComponent(err.Error(), time.Second*5)
				m.error = errorComponent
				return m, cmd
			}

			return m.parent, updateProfileNamesCmd
		default:
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			return m, cmd
		}
	case tea.WindowSizeMsg:
		m.root.SetSize(msg)
	case components.ErrorComponentTimeoutMsg:
		m.error = nil
	default:
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m renameProfile) View() string {
	inputView := lipgloss.NewStyle().Padding(1, 2).Render(m.input.View())

	if m.error != nil {
		return lipgloss.JoinVertical(lipgloss.Left, m.root.View(), m.title, m.error.View(), inputView)
	}

	return lipgloss.JoinVertical(lipgloss.Left, m.root.View(), m.title, inputView)
}
