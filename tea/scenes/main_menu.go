package scenes

import (
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rs/zerolog/log"
	"github.com/satisfactorymodding/ficsit-cli/tea/components"
	"github.com/satisfactorymodding/ficsit-cli/tea/utils"
)

var _ tea.Model = (*mainMenu)(nil)

type mainMenu struct {
	root  components.RootModel
	list  list.Model
	error *components.ErrorComponent
}

func NewMainMenu(root components.RootModel) tea.Model {
	model := mainMenu{
		root: root,
	}

	items := []list.Item{
		utils.SimpleItem[mainMenu]{
			ItemTitle: "Installations",
			Activate: func(msg tea.Msg, currentModel mainMenu) (tea.Model, tea.Cmd) {
				newModel := NewInstallations(root, currentModel)
				return newModel, newModel.Init()
			},
		},
		utils.SimpleItem[mainMenu]{
			ItemTitle: "Profiles",
			Activate: func(msg tea.Msg, currentModel mainMenu) (tea.Model, tea.Cmd) {
				newModel := NewProfiles(root, currentModel)
				return newModel, newModel.Init()
			},
		},
		utils.SimpleItem[mainMenu]{
			ItemTitle: "Mods",
			Activate: func(msg tea.Msg, currentModel mainMenu) (tea.Model, tea.Cmd) {
				newModel := NewMods(root, currentModel)
				return newModel, newModel.Init()
			},
		},
		utils.SimpleItem[mainMenu]{
			ItemTitle: "Apply Changes",
			Activate: func(msg tea.Msg, currentModel mainMenu) (tea.Model, tea.Cmd) {
				if err := root.GetGlobal().Save(); err != nil {
					log.Error().Err(err).Msg(ErrorFailedAddMod)
					errorComponent, cmd := components.NewErrorComponent(err.Error(), time.Second*5)
					currentModel.error = errorComponent
					return currentModel, cmd
				}

				newModel := NewApply(root, currentModel)
				return newModel, newModel.Init()
			},
		},
		utils.SimpleItem[mainMenu]{
			ItemTitle: "Save",
			Activate: func(msg tea.Msg, currentModel mainMenu) (tea.Model, tea.Cmd) {
				if err := root.GetGlobal().Save(); err != nil {
					log.Error().Err(err).Msg(ErrorFailedAddMod)
					errorComponent, cmd := components.NewErrorComponent(err.Error(), time.Second*5)
					currentModel.error = errorComponent
					return currentModel, cmd
				}
				return nil, nil
			},
		},
		utils.SimpleItem[mainMenu]{
			ItemTitle: "Exit",
			Activate: func(msg tea.Msg, currentModel mainMenu) (tea.Model, tea.Cmd) {
				return nil, tea.Quit
			},
		},
	}

	model.list = list.New(items, utils.NewItemDelegate(), root.Size().Width, root.Size().Height-root.Height())
	model.list.SetShowStatusBar(false)
	model.list.SetFilteringEnabled(false)
	model.list.Title = "Main Menu"
	model.list.Styles = utils.ListStyles
	model.list.DisableQuitKeybindings()

	return model
}

func (m mainMenu) Init() tea.Cmd {
	return nil
}

func (m mainMenu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case KeyControlC:
			return m, tea.Quit
		case "q":
			return m, tea.Quit
		case KeyEnter:
			i, ok := m.list.SelectedItem().(utils.SimpleItem[mainMenu])
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
			return m, nil
		default:
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}
	case tea.WindowSizeMsg:
		top, right, bottom, left := lipgloss.NewStyle().Margin(2, 2).GetMargin()
		m.list.SetSize(msg.Width-left-right, msg.Height-top-bottom)
		m.root.SetSize(msg)
	case components.ErrorComponentTimeoutMsg:
		m.error = nil
	}

	return m, nil
}

func (m mainMenu) View() string {
	if m.error != nil {
		err := (*m.error).View()
		m.list.SetSize(m.list.Width(), m.root.Size().Height-m.root.Height()-lipgloss.Height(err))
		return lipgloss.JoinVertical(lipgloss.Left, m.root.View(), err, m.list.View())
	}

	m.list.SetSize(m.list.Width(), m.root.Size().Height-m.root.Height())
	return lipgloss.JoinVertical(lipgloss.Left, m.root.View(), m.list.View())
}
