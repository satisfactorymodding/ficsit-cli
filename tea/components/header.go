package components

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/satisfactorymodding/ficsit-cli/tea/utils"
)

var _ tea.Model = (*headerComponent)(nil)

type headerComponent struct {
	root       RootModel
	labelStyle lipgloss.Style
}

func NewHeaderComponent(root RootModel) tea.Model {
	return headerComponent{
		root:       root,
		labelStyle: utils.LabelStyle,
	}
}

func (h headerComponent) Init() tea.Cmd {
	return nil
}

func (h headerComponent) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return h, nil
}

func (h headerComponent) View() string {
	out := h.labelStyle.Render("Installation: ")
	if h.root.GetCurrentInstallation() != nil {
		out += h.root.GetCurrentInstallation().Path
	} else {
		out += "None"
	}
	out += "\n"

	out += h.labelStyle.Render("Profile: ")
	if h.root.GetCurrentProfile() != nil {
		out += h.root.GetCurrentProfile().Name
	} else {
		out += "None"
	}

	return lipgloss.NewStyle().Margin(1, 0).Render(out)
}
