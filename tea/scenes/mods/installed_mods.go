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
	"github.com/rs/zerolog/log"

	"github.com/satisfactorymodding/ficsit-cli/ficsit"
	"github.com/satisfactorymodding/ficsit-cli/tea/components"
	"github.com/satisfactorymodding/ficsit-cli/tea/scenes/keys"
	"github.com/satisfactorymodding/ficsit-cli/tea/utils"
)

var _ tea.Model = (*modsList)(nil)

type installedModsList struct {
	list              list.Model
	sortFieldList     list.Model
	sortOrderList     list.Model
	root              components.RootModel
	parent            tea.Model
	items             chan listUpdate
	err               chan string
	error             *components.ErrorComponent
	sortingField      sortField
	sortingOrder      sortOrder
	showSortFieldList bool
	showSortOrderList bool
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
			key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "sort")),
			key.NewBinding(key.WithKeys("o"), key.WithHelp("o", "order")),
			key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
		}
	}
	l.AdditionalFullHelpKeys = l.AdditionalShortHelpKeys

	m := &installedModsList{
		root:         root,
		list:         l,
		parent:       parent,
		items:        make(chan listUpdate),
		sortingField: DefaultModSortingField,
		sortingOrder: DefaultModSortingOrder,
		err:          make(chan string),
	}

	m.sortFieldList = *m.newSortFieldsList(root)
	m.sortOrderList = *m.newSortOrderList(root)

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
		items[i] = utils.SimpleItemExtra[installedModsList, ficsit.ModsModsGetModsModsMod]{
			SimpleItem: utils.SimpleItem[installedModsList]{
				ItemTitle: reference,
				Activate: func(msg tea.Msg, currentModel installedModsList) (tea.Model, tea.Cmd) {
					return NewModMenu(m.root, currentModel, utils.Mod{
						Name:      r,
						Reference: r,
					}), nil
				},
			},
			Extra: ficsit.ModsModsGetModsModsMod{
				Name: r,
			},
		}
		i++
	}

	items = m.sortItems(items, m.sortingField, m.sortingOrder)

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

		items := make([]list.Item, 0)

		// create a pre-defined list of the mods in the profile, to then update
		// with modsmodsmodsmods info after
		for reference := range currentProfile.Mods {
			items = append(items, utils.SimpleItemExtra[installedModsList, ficsit.ModsModsGetModsModsMod]{
				SimpleItem: utils.SimpleItem[installedModsList]{
					ItemTitle: reference,
					Activate: func(msg tea.Msg, currentModel installedModsList) (tea.Model, tea.Cmd) {
						return NewModMenu(m.root, currentModel, utils.Mod{
							Name:      reference,
							Reference: reference,
						}), nil
					},
				},
				Extra: ficsit.ModsModsGetModsModsMod{
					Id:                reference,
					Name:              reference,
					Mod_reference:     reference,
					Views:             0,
					Downloads:         0,
					Popularity:        0,
					Hotness:           0,
					Created_at:        time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
					Last_version_date: time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
				},
			})
		}

		for i, mod := range mods.Mods.Mods {
			var index *int
			for exIndex, exItem := range items {
				if exItem.(utils.SimpleItemExtra[installedModsList, ficsit.ModsModsGetModsModsMod]).SimpleItem.ItemTitle == mod.Mod_reference {
					index = &exIndex
					break
				}
			}

			// Re-reference struct
			mod := mod
			item := utils.SimpleItemExtra[installedModsList, ficsit.ModsModsGetModsModsMod]{
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

			// if it already exists then replace it with the proper mod,
			// otherwise we will add it to the end
			if index != nil {
				items[*index] = item
			} else {
				items = append(items, item)
			}
		}

		items = m.sortItems(items, m.sortingField, m.sortingOrder)

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
		case "s":
			m.showSortFieldList = !m.showSortFieldList
			return m, nil
		case "o":
			m.showSortOrderList = !m.showSortOrderList
			return m, nil
		case keys.KeyControlC:
			return m, tea.Quit
		case "q":
			if m.showSortFieldList {
				m.showSortFieldList = false
				return m, nil
			}

			if m.showSortOrderList {
				m.showSortOrderList = false
				return m, nil
			}

			if m.parent != nil {
				m.parent.Update(m.root.Size())
				return m.parent, nil
			}
			return m, tea.Quit
		case keys.KeyEnter:
			if m.showSortFieldList {
				m.showSortFieldList = false
				i, ok := m.sortFieldList.SelectedItem().(utils.SimpleItem[installedModsList])
				if ok {
					return m.processActivation(i, msg)
				} else {
					log.Warn().Str("which", "field").Msg("could not cast selected item to simple item")
				}
				return m, nil
			}

			if m.showSortOrderList {
				m.showSortOrderList = false
				i, ok := m.sortOrderList.SelectedItem().(utils.SimpleItem[installedModsList])
				if ok {
					return m.processActivation(i, msg)
				} else {
					log.Warn().Str("which", "order").Msg("could not cast selected item to simple item")
				}
				return m, nil
			}

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

	if m.showSortFieldList {
		var cmd tea.Cmd
		m.sortFieldList, cmd = m.sortFieldList.Update(msg)
		return m, cmd
	} else if m.showSortOrderList {
		var cmd tea.Cmd
		m.sortOrderList, cmd = m.sortOrderList.Update(msg)
		return m, cmd
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m installedModsList) View() string {
	var bottomList list.Model
	if m.showSortFieldList {
		bottomList = m.sortFieldList
	} else if m.showSortOrderList {
		bottomList = m.sortOrderList
	} else {
		bottomList = m.list
	}

	bottomList.SetSize(bottomList.Width(), m.root.Size().Height-m.root.Height())
	return lipgloss.JoinVertical(lipgloss.Left, m.root.View(), bottomList.View())
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

func (m installedModsList) sortItems(items []list.Item, field sortField, direction sortOrder) []list.Item {
	sortedItems := make([]list.Item, len(items))
	copy(sortedItems, items)

	switch field {
	case sortFieldLastVersionDate:
		sort.Slice(sortedItems, func(i, j int) bool {
			a := sortedItems[i].(utils.SimpleItemExtra[installedModsList, ficsit.ModsModsGetModsModsMod])
			b := sortedItems[j].(utils.SimpleItemExtra[installedModsList, ficsit.ModsModsGetModsModsMod])
			return ascDesc(direction, a.Extra.Last_version_date.Before(b.Extra.Last_version_date))
		})
	case sortFieldCreatedAt:
		sort.Slice(sortedItems, func(i, j int) bool {
			a := sortedItems[i].(utils.SimpleItemExtra[installedModsList, ficsit.ModsModsGetModsModsMod])
			b := sortedItems[j].(utils.SimpleItemExtra[installedModsList, ficsit.ModsModsGetModsModsMod])
			return ascDesc(direction, a.Extra.Created_at.Before(b.Extra.Created_at))
		})
	case sortFieldName:
		sort.Slice(sortedItems, func(i, j int) bool {
			a := sortedItems[i].(utils.SimpleItemExtra[installedModsList, ficsit.ModsModsGetModsModsMod])
			b := sortedItems[j].(utils.SimpleItemExtra[installedModsList, ficsit.ModsModsGetModsModsMod])
			return ascDesc(direction, a.Extra.Name < b.Extra.Name)
		})
	case sortFieldDownloads:
		sort.Slice(sortedItems, func(i, j int) bool {
			a := sortedItems[i].(utils.SimpleItemExtra[installedModsList, ficsit.ModsModsGetModsModsMod])
			b := sortedItems[j].(utils.SimpleItemExtra[installedModsList, ficsit.ModsModsGetModsModsMod])
			return ascDesc(direction, a.Extra.Downloads < b.Extra.Downloads)
		})
	case sortFieldViews:
		sort.Slice(sortedItems, func(i, j int) bool {
			a := sortedItems[i].(utils.SimpleItemExtra[installedModsList, ficsit.ModsModsGetModsModsMod])
			b := sortedItems[j].(utils.SimpleItemExtra[installedModsList, ficsit.ModsModsGetModsModsMod])
			return ascDesc(direction, a.Extra.Views < b.Extra.Views)
		})
	case sortFieldPopularity:
		sort.Slice(sortedItems, func(i, j int) bool {
			a := sortedItems[i].(utils.SimpleItemExtra[installedModsList, ficsit.ModsModsGetModsModsMod])
			b := sortedItems[j].(utils.SimpleItemExtra[installedModsList, ficsit.ModsModsGetModsModsMod])
			return ascDesc(direction, a.Extra.Popularity < b.Extra.Popularity)
		})
	case sortFieldHotness:
		sort.Slice(sortedItems, func(i, j int) bool {
			a := sortedItems[i].(utils.SimpleItemExtra[installedModsList, ficsit.ModsModsGetModsModsMod])
			b := sortedItems[j].(utils.SimpleItemExtra[installedModsList, ficsit.ModsModsGetModsModsMod])
			return ascDesc(direction, a.Extra.Hotness < b.Extra.Hotness)
		})
	}

	return sortedItems
}

func (m installedModsList) newSortFieldsList(root components.RootModel) *list.Model {
	sortFieldList := list.New([]list.Item{
		utils.SimpleItem[installedModsList]{
			ItemTitle: "Name",
			Activate: func(msg tea.Msg, m installedModsList) (tea.Model, tea.Cmd) {
				m.sortingField = sortFieldName
				cmd := m.list.SetItems(m.sortItems(m.list.Items(), m.sortingField, m.sortingOrder))
				m.list.ResetSelected()
				return m, cmd
			},
		},
		utils.SimpleItem[installedModsList]{
			ItemTitle: "Last Version Date",
			Activate: func(msg tea.Msg, m installedModsList) (tea.Model, tea.Cmd) {
				m.sortingField = sortFieldLastVersionDate
				cmd := m.list.SetItems(m.sortItems(m.list.Items(), m.sortingField, m.sortingOrder))
				m.list.ResetSelected()
				return m, cmd
			},
		},
		utils.SimpleItem[installedModsList]{
			ItemTitle: "Creation Date",
			Activate: func(msg tea.Msg, m installedModsList) (tea.Model, tea.Cmd) {
				m.sortingField = sortFieldCreatedAt
				cmd := m.list.SetItems(m.sortItems(m.list.Items(), m.sortingField, m.sortingOrder))
				m.list.ResetSelected()
				return m, cmd
			},
		},
		utils.SimpleItem[installedModsList]{
			ItemTitle: "Downloads",
			Activate: func(msg tea.Msg, m installedModsList) (tea.Model, tea.Cmd) {
				m.sortingField = sortFieldDownloads
				cmd := m.list.SetItems(m.sortItems(m.list.Items(), m.sortingField, m.sortingOrder))
				m.list.ResetSelected()
				return m, cmd
			},
		},
		utils.SimpleItem[installedModsList]{
			ItemTitle: "Views",
			Activate: func(msg tea.Msg, m installedModsList) (tea.Model, tea.Cmd) {
				m.sortingField = sortFieldViews
				cmd := m.list.SetItems(m.sortItems(m.list.Items(), m.sortingField, m.sortingOrder))
				m.list.ResetSelected()
				return m, cmd
			},
		},
		utils.SimpleItem[installedModsList]{
			ItemTitle: "Popularity (recent downloads)",
			Activate: func(msg tea.Msg, m installedModsList) (tea.Model, tea.Cmd) {
				m.sortingField = sortFieldPopularity
				cmd := m.list.SetItems(m.sortItems(m.list.Items(), m.sortingField, m.sortingOrder))
				m.list.ResetSelected()
				return m, cmd
			},
		},
		utils.SimpleItem[installedModsList]{
			ItemTitle: "Hotness (recent views)",
			Activate: func(msg tea.Msg, m installedModsList) (tea.Model, tea.Cmd) {
				m.sortingField = sortFieldHotness
				cmd := m.list.SetItems(m.sortItems(m.list.Items(), m.sortingField, m.sortingOrder))
				m.list.ResetSelected()
				return m, cmd
			},
		},
	}, utils.NewItemDelegate(), root.Size().Width, root.Size().Height-root.Height())
	sortFieldList.SetShowStatusBar(true)
	sortFieldList.SetShowFilter(false)
	sortFieldList.SetFilteringEnabled(false)
	sortFieldList.Title = m.list.Title
	sortFieldList.Styles = utils.ListStyles
	sortFieldList.SetSize(m.list.Width(), m.list.Height())
	sortFieldList.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
		}
	}

	return &sortFieldList
}

func (m installedModsList) newSortOrderList(root components.RootModel) *list.Model {
	sortOrderList := list.New([]list.Item{
		utils.SimpleItem[installedModsList]{
			ItemTitle: "Ascending",
			Activate: func(msg tea.Msg, m installedModsList) (tea.Model, tea.Cmd) {
				m.sortingOrder = sortOrderAsc
				cmd := m.list.SetItems(m.sortItems(m.list.Items(), m.sortingField, m.sortingOrder))
				m.list.ResetSelected()
				return m, cmd
			},
		},
		utils.SimpleItem[installedModsList]{
			ItemTitle: "Descending",
			Activate: func(msg tea.Msg, m installedModsList) (tea.Model, tea.Cmd) {
				m.sortingOrder = sortOrderDesc
				cmd := m.list.SetItems(m.sortItems(m.list.Items(), m.sortingField, m.sortingOrder))
				m.list.ResetSelected()
				return m, cmd
			},
		},
	}, utils.NewItemDelegate(), root.Size().Width, root.Size().Height-root.Height())
	sortOrderList.SetShowStatusBar(true)
	sortOrderList.SetShowFilter(false)
	sortOrderList.SetFilteringEnabled(false)
	sortOrderList.Title = m.list.Title
	sortOrderList.Styles = utils.ListStyles
	sortOrderList.SetSize(m.list.Width(), m.list.Height())
	sortOrderList.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
		}
	}

	return &sortOrderList
}
