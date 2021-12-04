package scenes

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/davecgh/go-spew/spew"
	"github.com/rs/zerolog/log"
	"github.com/satisfactorymodding/ficsit-cli/tea/components"
	"github.com/satisfactorymodding/ficsit-cli/tea/utils"
)

var _ tea.Model = (*mainMenu)(nil)

type mainMenu struct {
	root components.RootModel
	list list.Model
}

func NewMainMenu(root components.RootModel) tea.Model {
	model := mainMenu{
		root: root,
	}

	items := []list.Item{
		utils.SimpleItem{
			ItemTitle: "Installations",
			Activate: func(msg tea.Msg, currentModel tea.Model) (tea.Model, tea.Cmd) {
				newModel := NewInstallations(root, currentModel)
				return newModel, newModel.Init()
			},
		},
		utils.SimpleItem{
			ItemTitle: "Profiles",
			Activate: func(msg tea.Msg, currentModel tea.Model) (tea.Model, tea.Cmd) {
				newModel := NewProfiles(root, currentModel)
				return newModel, newModel.Init()
			},
		},
		utils.SimpleItem{
			ItemTitle: "Mods",
			Activate: func(msg tea.Msg, currentModel tea.Model) (tea.Model, tea.Cmd) {
				newModel := NewMods(root, currentModel)
				return newModel, newModel.Init()
			},
		},
		utils.SimpleItem{
			ItemTitle: "Apply Changes",
			Activate: func(msg tea.Msg, currentModel tea.Model) (tea.Model, tea.Cmd) {
				// TODO Apply changes to all changed profiles
				return nil, nil
			},
		},
		utils.SimpleItem{
			ItemTitle: "Save",
			Activate: func(msg tea.Msg, currentModel tea.Model) (tea.Model, tea.Cmd) {
				if err := root.GetGlobal().Save(); err != nil {
					panic(err) // TODO Handle Error
				}
				return nil, nil
			},
		},
		utils.SimpleItem{
			ItemTitle: "Exit",
			Activate: func(msg tea.Msg, currentModel tea.Model) (tea.Model, tea.Cmd) {
				newModel := NewExitMenu(root)
				return newModel, newModel.Init()
			},
		},
	}

	model.list = list.NewModel(items, utils.NewItemDelegate(), root.Size().Width, root.Size().Height-root.Height())
	model.list.SetShowStatusBar(false)
	model.list.SetFilteringEnabled(false)
	model.list.Title = "Main Menu"
	model.list.Styles = utils.ListStyles
	model.list.SetSize(model.list.Width(), model.list.Height())

	return model
}

func (m mainMenu) Init() tea.Cmd {
	return nil
}

func (m mainMenu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	log.Warn().Msg(spew.Sdump(msg))
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case KeyControlC:
			return m, tea.Quit
		case "q":
			newModel := NewExitMenu(m.root)
			return newModel, newModel.Init()
		case KeyEnter:
			i, ok := m.list.SelectedItem().(utils.SimpleItem)
			if ok {
				if i.Activate != nil {
					newModel, cmd := i.Activate(msg, m)
					if newModel != nil || cmd != nil {
						if newModel == nil {
							newModel = m
						}
						return newModel, cmd
					}
					return m, nil
				}
			}
			return m, tea.Quit
		default:
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}
	case tea.WindowSizeMsg:
		top, right, bottom, left := lipgloss.NewStyle().Margin(2, 2).GetMargin()
		m.list.SetSize(msg.Width-left-right, msg.Height-top-bottom)
		m.root.SetSize(msg)
	}

	return m, nil
}

func (m mainMenu) View() string {
	return lipgloss.JoinVertical(lipgloss.Left, m.root.View(), m.list.View())
}
