package mods

import (
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/satisfactorymodding/ficsit-cli/tea/components"
	"github.com/satisfactorymodding/ficsit-cli/tea/scenes/keys"
	"github.com/satisfactorymodding/ficsit-cli/tea/utils"
)

var _ tea.Model = (*modSemver)(nil)

type modSemver struct {
	input  textinput.Model
	root   components.RootModel
	parent tea.Model
	error  *components.ErrorComponent
	mod    utils.Mod
	title  string
}

func NewModSemver(root components.RootModel, parent tea.Model, mod utils.Mod) tea.Model {
	model := modSemver{
		root:   root,
		parent: parent,
		input:  textinput.New(),
		title:  lipgloss.NewStyle().Padding(0, 2).Render(utils.TitleStyle.Render(mod.Name)),
		mod:    mod,
	}

	model.input.Placeholder = ">=1.2.3"
	model.input.Focus()
	model.input.Width = root.Size().Width

	return model
}

func (m modSemver) Init() tea.Cmd {
	return textinput.Blink
}

func (m modSemver) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case keys.KeyControlC:
			return m, tea.Quit
		case keys.KeyEscape:
			return m.parent, nil
		case keys.KeyEnter:
			err := m.root.GetCurrentProfile().AddMod(m.mod.Reference, m.input.Value())
			if err != nil {
				errorComponent, cmd := components.NewErrorComponent(err.Error(), time.Second*5)
				m.error = errorComponent
				return m, cmd
			}
			return m.parent, nil
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

func (m modSemver) View() string {
	inputView := lipgloss.NewStyle().Padding(1, 2).Render(m.input.View())

	if m.error != nil {
		return lipgloss.JoinVertical(lipgloss.Left, m.root.View(), m.title, m.error.View(), inputView)
	}

	return lipgloss.JoinVertical(lipgloss.Left, m.root.View(), m.title, inputView)
}
