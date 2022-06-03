package scenes

import (
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/satisfactorymodding/ficsit-cli/tea/components"
	"github.com/satisfactorymodding/ficsit-cli/tea/utils"
)

var _ tea.Model = (*newProfile)(nil)

type newProfile struct {
	root   components.RootModel
	parent tea.Model
	input  textinput.Model
	title  string
	error  *components.ErrorComponent
}

func NewNewProfile(root components.RootModel, parent tea.Model) tea.Model {
	model := newProfile{
		root:   root,
		parent: parent,
		input:  textinput.New(),
		title:  utils.NonListTitleStyle.Render("New Profile"),
	}

	model.input.Focus()
	model.input.Width = root.Size().Width

	return model
}

func (m newProfile) Init() tea.Cmd {
	return nil
}

func (m newProfile) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case KeyControlC:
			return m, tea.Quit
		case KeyEscape:
			return m.parent, nil
		case KeyEnter:
			if _, err := m.root.GetGlobal().Profiles.AddProfile(m.input.Value()); err != nil {
				errorComponent, cmd := components.NewErrorComponent(err.Error(), time.Second*5)
				m.error = errorComponent
				return m, cmd
			}

			return m.parent, updateProfileListCmd
		default:
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			return m, cmd
		}
	case tea.WindowSizeMsg:
		m.root.SetSize(msg)
	case components.ErrorComponentTimeoutMsg:
		m.error = nil
	}

	return m, nil
}

func (m newProfile) View() string {
	inputView := lipgloss.NewStyle().Padding(1, 2).Render(m.input.View())

	if m.error != nil {
		return lipgloss.JoinVertical(lipgloss.Left, m.root.View(), m.title, (*m.error).View(), inputView)
	}

	return lipgloss.JoinVertical(lipgloss.Left, m.root.View(), m.title, inputView)
}
