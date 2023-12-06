package mods

import (
	"context"
	"sort"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/satisfactorymodding/ficsit-cli/ficsit"
	"github.com/satisfactorymodding/ficsit-cli/tea/components"
	"github.com/satisfactorymodding/ficsit-cli/tea/scenes/keys"
	"github.com/satisfactorymodding/ficsit-cli/tea/utils"
)

var _ tea.Model = (*installedModsList)(nil)

type installedModsList struct {
	root   components.RootModel
	list   list.Model
	parent tea.Model
	items  chan listUpdate

	err   chan string
	error *components.ErrorComponent
}

func NewInstalledMods(root components.RootModel, parent tea.Model) tea.Model {
	currentProfile := root.GetCurrentProfile()
	if currentProfile == nil {
		return parent
	}

	l := list.New([]list.Item{}, utils.NewItemDelegate(), root.Size().Width, root.Size().Height-root.Height())
	l.SetShowStatusBar(true)
	l.SetShowFilter(true)
	l.SetFilteringEnabled(true)
	l.SetSpinner(spinner.MiniDot)
	l.Title = "Installed Mods"
	l.Styles = utils.ListStyles
	l.SetSize(l.Width(), l.Height())
	l.KeyMap.Quit.SetHelp("q", "back")
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
		}
	}

	m := &installedModsList{
		root:   root,
		list:   l,
		parent: parent,
		items:  make(chan listUpdate),
		err:    make(chan string),
	}

	return m
}

func (m installedModsList) Init() tea.Cmd {
	m.LoadModData()
	return utils.Ticker()
}

func (m installedModsList) LoadModData() {
	currentProfile := m.root.GetCurrentProfile()
	if currentProfile == nil {
		return
	}

	items := make([]list.Item, len(currentProfile.Mods))
	i := 0
	for reference := range currentProfile.Mods {
		r := reference
		items[i] = utils.SimpleItem[installedModsList]{
			ItemTitle: reference,
			Activate: func(msg tea.Msg, currentModel installedModsList) (tea.Model, tea.Cmd) {
				return NewModMenu(m.root, currentModel, utils.Mod{
					Name:      r,
					Reference: r,
				}), nil
			},
		}
		i++
	}

	sort.Slice(items, func(i, j int) bool {
		a := items[i].(utils.SimpleItem[installedModsList])
		b := items[j].(utils.SimpleItem[installedModsList])
		return ascDesc(sortOrderDesc, a.ItemTitle < b.ItemTitle)
	})

	go func() {
		if len(currentProfile.Mods) == 0 {
			m.items <- listUpdate{
				Items: items,
				Done:  true,
			}
			return
			// Continuing past this point would load info about mods we don't have installed
		}

		references := make([]string, len(currentProfile.Mods))
		i := 0
		for reference := range currentProfile.Mods {
			references[i] = reference
			i++
		}

		mods, err := ficsit.Mods(context.TODO(), m.root.GetAPIClient(), ficsit.ModFilter{
			References: references,
		})
		if err != nil {
			m.err <- err.Error()
			return
		}

		if len(mods.Mods.Mods) == 0 {
			return
		}

		items := make([]list.Item, len(mods.Mods.Mods))
		for i, mod := range mods.Mods.Mods {
			// Re-reference struct
			mod := mod
			items[i] = utils.SimpleItemExtra[installedModsList, ficsit.ModsModsGetModsModsMod]{
				SimpleItem: utils.SimpleItem[installedModsList]{
					ItemTitle: mods.Mods.Mods[i].Name,
					Activate: func(msg tea.Msg, currentModel installedModsList) (tea.Model, tea.Cmd) {
						return NewModMenu(m.root, currentModel, utils.Mod{
							Name:      mod.Name,
							Reference: mod.Mod_reference,
						}), nil
					},
				},
				Extra: mod,
			}
		}

		sort.Slice(items, func(i, j int) bool {
			a := items[i].(utils.SimpleItemExtra[installedModsList, ficsit.ModsModsGetModsModsMod])
			b := items[j].(utils.SimpleItemExtra[installedModsList, ficsit.ModsModsGetModsModsMod])
			return ascDesc(sortOrderDesc, a.Extra.Mod_reference < b.Extra.Mod_reference)
		})

		m.items <- listUpdate{
			Items: items,
			Done:  true,
		}
	}()

	go func() {
		m.items <- listUpdate{
			Items: items,
			Done:  false,
		}
	}()
}

func (m installedModsList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.list.SettingFilter() {
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}

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
			i, ok := m.list.SelectedItem().(utils.SimpleItem[installedModsList])
			if ok {
				return m.processActivation(i, msg)
			}
			i2, ok := m.list.SelectedItem().(utils.SimpleItemExtra[installedModsList, ficsit.ModsModsGetModsModsMod])
			if ok {
				return m.processActivation(i2.SimpleItem, msg)
			}
			return m, nil
		}
	case tea.WindowSizeMsg:
		top, right, bottom, left := lipgloss.NewStyle().Margin(m.root.Height(), 2, 0).GetMargin()
		m.list.SetSize(msg.Width-left-right, msg.Height-top-bottom)
		m.root.SetSize(msg)
	case utils.TickMsg:
		select {
		case items := <-m.items:
			cmd := m.list.SetItems(items.Items)
			if items.Done {
				m.list.StopSpinner()
				return m, cmd
			}
			return m, tea.Batch(utils.Ticker(), cmd)
		case err := <-m.err:
			errorComponent, cmd := components.NewErrorComponent(err, time.Second*5)
			m.error = errorComponent
			return m, cmd
		default:
			start := m.list.StartSpinner()
			return m, tea.Batch(utils.Ticker(), start)
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m installedModsList) View() string {
	m.list.SetSize(m.list.Width(), m.root.Size().Height-m.root.Height())
	return lipgloss.JoinVertical(lipgloss.Left, m.root.View(), m.list.View())
}

func (m installedModsList) processActivation(item utils.SimpleItem[installedModsList], msg tea.Msg) (tea.Model, tea.Cmd) {
	if item.Activate != nil {
		newModel, cmd := item.Activate(msg, m)
		if newModel != nil || cmd != nil {
			if newModel == nil {
				newModel = m
			}
			return newModel, cmd
		}
		return m, nil
	}
	return m, nil
}
