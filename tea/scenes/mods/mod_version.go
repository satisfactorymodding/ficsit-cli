package mods

import (
	"log/slog"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/satisfactorymodding/ficsit-cli/tea/components"
	"github.com/satisfactorymodding/ficsit-cli/tea/scenes/errors"
	"github.com/satisfactorymodding/ficsit-cli/tea/scenes/keys"
	"github.com/satisfactorymodding/ficsit-cli/tea/utils"
)

var _ tea.Model = (*modVersionMenu)(nil)

type modVersionMenu struct {
	root   components.RootModel
	list   list.Model
	parent tea.Model
}

func NewModVersion(root components.RootModel, parent tea.Model, mod utils.Mod) tea.Model {
	model := modVersionMenu{
		root:   root,
		parent: parent,
	}

	items := []list.Item{
		utils.SimpleItem[modVersionMenu]{
			ItemTitle: "Select Version",
			Activate: func(msg tea.Msg, currentModel modVersionMenu) (tea.Model, tea.Cmd) {
				newModel := NewModVersionList(root, currentModel.parent, mod)
				return newModel, newModel.Init()
			},
		},
		utils.SimpleItem[modVersionMenu]{
			ItemTitle: "Enter Custom SemVer",
			Activate: func(msg tea.Msg, currentModel modVersionMenu) (tea.Model, tea.Cmd) {
				newModel := NewModSemver(root, currentModel.parent, mod)
				return newModel, newModel.Init()
			},
		},
	}

	if root.GetCurrentProfile().HasMod(mod.Reference) {
		items = append([]list.Item{
			utils.SimpleItem[modVersionMenu]{
				ItemTitle: "Latest",
				Activate: func(msg tea.Msg, currentModel modVersionMenu) (tea.Model, tea.Cmd) {
					err := root.GetCurrentProfile().AddMod(mod.Reference, ">=0.0.0")
					if err != nil {
						slog.Error(errors.ErrorFailedAddMod, slog.Any("err", err))
						cmd := currentModel.list.NewStatusMessage(errors.ErrorFailedAddMod)
						return currentModel, cmd
					}
					return currentModel.parent, nil
				},
			},
		}, items...)
	}

	model.list = list.New(items, utils.NewItemDelegate(), root.Size().Width, root.Size().Height-root.Height())
	model.list.SetShowStatusBar(false)
	model.list.SetFilteringEnabled(false)
	model.list.Title = mod.Name
	model.list.Styles = utils.ListStyles
	model.list.SetSize(model.list.Width(), model.list.Height())
	model.list.StatusMessageLifetime = time.Second * 3
	model.list.KeyMap.Quit.SetHelp("q", "back")
	model.list.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
		}
	}

	return model
}

func (m modVersionMenu) Init() tea.Cmd {
	return nil
}

func (m modVersionMenu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case keys.KeyControlC:
			return m, tea.Quit
		case "q":
			if m.parent != nil {
				m.parent.Update(m.root.Size())
				return m.parent, nil
			}
			return m, tea.Quit
		case keys.KeyEnter:
			i, ok := m.list.SelectedItem().(utils.SimpleItem[modVersionMenu])
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
	}

	return m, nil
}

func (m modVersionMenu) View() string {
	return lipgloss.JoinVertical(lipgloss.Left, m.root.View(), m.list.View())
}
