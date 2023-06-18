package profile

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/satisfactorymodding/ficsit-cli/cli"
	"github.com/satisfactorymodding/ficsit-cli/tea/components"
	"github.com/satisfactorymodding/ficsit-cli/tea/scenes/keys"
	"github.com/satisfactorymodding/ficsit-cli/tea/utils"
)

var _ tea.Model = (*profile)(nil)

type profile struct {
	list       list.Model
	root       components.RootModel
	parent     tea.Model
	profile    *cli.Profile
	error      *components.ErrorComponent
	hadRenamed bool
}

func NewProfile(root components.RootModel, parent tea.Model, profileData *cli.Profile) tea.Model {
	model := profile{
		root:    root,
		parent:  parent,
		profile: profileData,
	}

	items := []list.Item{
		utils.SimpleItem[profile]{
			ItemTitle: "Select",
			Activate: func(msg tea.Msg, currentModel profile) (tea.Model, tea.Cmd) {
				if err := root.SetCurrentProfile(profileData); err != nil {
					errorComponent, cmd := components.NewErrorComponent(err.Error(), time.Second*5)
					currentModel.error = errorComponent
					return currentModel, cmd
				}

				return currentModel.parent, nil
			},
		},
	}

	if profileData.Name != cli.DefaultProfileName {
		items = append(items,
			utils.SimpleItem[profile]{
				ItemTitle: "Rename",
				Activate: func(msg tea.Msg, currentModel profile) (tea.Model, tea.Cmd) {
					newModel := NewRenameProfile(root, currentModel, profileData)
					return newModel, newModel.Init()
				},
			},
			utils.SimpleItem[profile]{
				ItemTitle: "Delete",
				Activate: func(msg tea.Msg, currentModel profile) (tea.Model, tea.Cmd) {
					if err := root.GetGlobal().Profiles.DeleteProfile(profileData.Name); err != nil {
						errorComponent, cmd := components.NewErrorComponent(err.Error(), time.Second*5)
						currentModel.error = errorComponent
						return currentModel, cmd
					}

					return currentModel.parent, updateProfileListCmd
				},
			},
		)
	}

	model.list = list.New(items, utils.NewItemDelegate(), root.Size().Width, root.Size().Height-root.Height())
	model.list.SetShowStatusBar(false)
	model.list.SetFilteringEnabled(false)
	model.list.Title = fmt.Sprintf("Profile: %s", profileData.Name)
	model.list.Styles = utils.ListStyles
	model.list.SetSize(model.list.Width(), model.list.Height())
	model.list.StatusMessageLifetime = time.Second * 3
	model.list.KeyMap.Quit.SetHelp("q", "back")

	return model
}

func (m profile) Init() tea.Cmd {
	return nil
}

func (m profile) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case keys.KeyControlC:
			return m, tea.Quit
		case "q":
			if m.parent != nil {
				m.parent.Update(m.root.Size())

				if m.hadRenamed {
					return m.parent, updateProfileNamesCmd
				}

				return m.parent, nil
			}
			return m, nil
		case keys.KeyEnter:
			i, ok := m.list.SelectedItem().(utils.SimpleItem[profile])
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
	case updateProfileNames:
		m.hadRenamed = true
		m.list.Title = fmt.Sprintf("Profile: %s", m.profile.Name)
	case components.ErrorComponentTimeoutMsg:
		m.error = nil
	}

	return m, nil
}

func (m profile) View() string {
	if m.error != nil {
		err := m.error.View()
		m.list.SetSize(m.list.Width(), m.root.Size().Height-m.root.Height()-lipgloss.Height(err))
		return lipgloss.JoinVertical(lipgloss.Left, m.root.View(), err, m.list.View())
	}

	m.list.SetSize(m.list.Width(), m.root.Size().Height-m.root.Height())
	return lipgloss.JoinVertical(lipgloss.Left, m.root.View(), m.list.View())
}
