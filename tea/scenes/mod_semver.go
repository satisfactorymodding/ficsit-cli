package scenes

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/satisfactorymodding/ficsit-cli/tea/components"
	"github.com/satisfactorymodding/ficsit-cli/tea/utils"
)

var _ tea.Model = (*modSemver)(nil)

type modSemver struct {
	root   components.RootModel
	parent tea.Model
	input  textinput.Model
	title  string
	mod    utils.Mod
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
	return nil
}

func (m modSemver) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case KeyControlC:
			return m, tea.Quit
		case KeyEscape:
			return m.parent, nil
		case KeyEnter:
			err := m.root.GetCurrentProfile().AddMod(m.mod.Reference, m.input.Value())
			if err != nil {
				panic(err) // TODO Handle Error
			}
			return m.parent, nil
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

func (m modSemver) View() string {
	inputView := lipgloss.NewStyle().Padding(1, 2).Render(m.input.View())
	return lipgloss.JoinVertical(lipgloss.Left, m.root.View(), m.title, inputView)
}
