package scenes

import (
	"context"
	"sort"

	"github.com/charmbracelet/bubbles/key"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/satisfactorymodding/ficsit-cli/ficsit"
	"github.com/satisfactorymodding/ficsit-cli/tea/components"
	"github.com/satisfactorymodding/ficsit-cli/tea/utils"
)

var _ tea.Model = (*modsList)(nil)

type sortOrder string

const (
	sortOrderAsc  sortOrder = "asc"
	sortOrderDesc sortOrder = "desc"
)

const modsTitle = "Mods"

type modsList struct {
	root   components.RootModel
	list   list.Model
	parent tea.Model
	items  chan []list.Item

	sortingField string
	sortingOrder sortOrder

	showSortFieldList bool
	sortFieldList     list.Model

	showSortOrderList bool
	sortOrderList     list.Model
}

func NewMods(root components.RootModel, parent tea.Model) tea.Model {
	// TODO Color mods that are installed in current profile
	l := list.NewModel([]list.Item{}, utils.NewItemDelegate(), root.Size().Width, root.Size().Height-root.Height())
	l.SetShowStatusBar(true)
	l.SetShowFilter(true)
	l.SetFilteringEnabled(true)
	l.SetSpinner(spinner.MiniDot)
	l.Title = modsTitle
	l.Styles = utils.ListStyles
	l.SetSize(l.Width(), l.Height())
	l.KeyMap.Quit.SetHelp("q", "back")
	l.DisableQuitKeybindings()

	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(key.WithHelp("s", "sort")),
			key.NewBinding(key.WithHelp("o", "order")),
		}
	}

	l.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(key.WithHelp("s", "sort")),
			key.NewBinding(key.WithHelp("o", "order")),
		}
	}

	sortFieldList := list.NewModel([]list.Item{
		utils.SimpleItem{
			ItemTitle: "Name",
			Activate: func(msg tea.Msg, currentModel tea.Model) (tea.Model, tea.Cmd) {
				m := currentModel.(modsList)
				m.sortingField = "name"
				cmd := m.list.SetItems(sortItems(m.list.Items(), m.sortingField, m.sortingOrder))
				m.list.ResetSelected()
				return m, cmd
			},
		},
		utils.SimpleItem{
			ItemTitle: "Last Version Date",
			Activate: func(msg tea.Msg, currentModel tea.Model) (tea.Model, tea.Cmd) {
				m := currentModel.(modsList)
				m.sortingField = "last_version_date"
				cmd := m.list.SetItems(sortItems(m.list.Items(), m.sortingField, m.sortingOrder))
				m.list.ResetSelected()
				return m, cmd
			},
		},
		utils.SimpleItem{
			ItemTitle: "Creation Date",
			Activate: func(msg tea.Msg, currentModel tea.Model) (tea.Model, tea.Cmd) {
				m := currentModel.(modsList)
				m.sortingField = "created_at"
				cmd := m.list.SetItems(sortItems(m.list.Items(), m.sortingField, m.sortingOrder))
				m.list.ResetSelected()
				return m, cmd
			},
		},
	}, utils.NewItemDelegate(), root.Size().Width, root.Size().Height-root.Height())
	sortFieldList.SetShowStatusBar(true)
	sortFieldList.SetShowFilter(false)
	sortFieldList.SetFilteringEnabled(false)
	sortFieldList.Title = modsTitle
	sortFieldList.Styles = utils.ListStyles
	sortFieldList.SetSize(l.Width(), l.Height())
	sortFieldList.KeyMap.Quit.SetHelp("q", "back")
	sortFieldList.DisableQuitKeybindings()

	sortOrderList := list.NewModel([]list.Item{
		utils.SimpleItem{
			ItemTitle: "Ascending",
			Activate: func(msg tea.Msg, currentModel tea.Model) (tea.Model, tea.Cmd) {
				m := currentModel.(modsList)
				m.sortingOrder = sortOrderAsc
				cmd := m.list.SetItems(sortItems(m.list.Items(), m.sortingField, m.sortingOrder))
				m.list.ResetSelected()
				return m, cmd
			},
		},
		utils.SimpleItem{
			ItemTitle: "Descending",
			Activate: func(msg tea.Msg, currentModel tea.Model) (tea.Model, tea.Cmd) {
				m := currentModel.(modsList)
				m.sortingOrder = sortOrderDesc
				cmd := m.list.SetItems(sortItems(m.list.Items(), m.sortingField, m.sortingOrder))
				m.list.ResetSelected()
				return m, cmd
			},
		},
	}, utils.NewItemDelegate(), root.Size().Width, root.Size().Height-root.Height())
	sortOrderList.SetShowStatusBar(true)
	sortOrderList.SetShowFilter(false)
	sortOrderList.SetFilteringEnabled(false)
	sortOrderList.Title = modsTitle
	sortOrderList.Styles = utils.ListStyles
	sortOrderList.SetSize(l.Width(), l.Height())
	sortOrderList.KeyMap.Quit.SetHelp("q", "back")
	sortOrderList.DisableQuitKeybindings()

	m := &modsList{
		root:          root,
		list:          l,
		parent:        parent,
		items:         make(chan []list.Item),
		sortingField:  "last_version_date",
		sortingOrder:  sortOrderDesc,
		sortFieldList: sortFieldList,
		sortOrderList: sortOrderList,
	}

	go func() {
		items := make([]list.Item, 0)
		allMods := make([]ficsit.ModsGetModsModsMod, 0)
		offset := 0
		for {
			mods, err := ficsit.Mods(context.TODO(), root.GetAPIClient(), ficsit.ModFilter{
				Limit:    100,
				Offset:   offset,
				Order_by: ficsit.ModFieldsLastVersionDate,
				Order:    ficsit.OrderDesc,
			})

			if err != nil {
				panic(err) // TODO Handle Error
			}

			if len(mods.GetMods.Mods) == 0 {
				break
			}

			allMods = append(allMods, mods.GetMods.Mods...)

			for i := 0; i < len(mods.GetMods.Mods); i++ {
				currentOffset := offset
				currentI := i
				items = append(items, utils.SimpleItem{
					ItemTitle: mods.GetMods.Mods[i].Name,
					Activate: func(msg tea.Msg, currentModel tea.Model) (tea.Model, tea.Cmd) {
						mod := allMods[currentOffset+currentI]
						return NewModMenu(root, currentModel, utils.Mod{
							Name:      mod.Name,
							ID:        mod.Id,
							Reference: mod.Mod_reference,
						}), nil
					},
					Extra: allMods[currentOffset+currentI],
				})
			}

			offset += len(mods.GetMods.Mods)
		}

		m.items <- items
	}()

	return m
}

func (m modsList) Init() tea.Cmd {
	return utils.Ticker()
}

func (m modsList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// List enables its own keybindings when they were previously disabled
	m.list.DisableQuitKeybindings()

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
		case KeyControlC:
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
		case KeyEnter:
			if m.showSortFieldList {
				m.showSortFieldList = false
				i, ok := m.sortFieldList.SelectedItem().(utils.SimpleItem)
				if ok {
					return m.processActivation(i, msg)
				}
				return m, nil
			}

			if m.showSortOrderList {
				m.showSortOrderList = false
				i, ok := m.sortOrderList.SelectedItem().(utils.SimpleItem)
				if ok {
					return m.processActivation(i, msg)
				}
				return m, nil
			}

			i, ok := m.list.SelectedItem().(utils.SimpleItem)
			if ok {
				return m.processActivation(i, msg)
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
			m.list.StopSpinner()
			cmd := m.list.SetItems(items)
			// Done to refresh keymap
			m.list.SetFilteringEnabled(m.list.FilteringEnabled())
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

func (m modsList) View() string {
	var bottom string
	if m.showSortFieldList {
		bottom = m.sortFieldList.View()
	} else if m.showSortOrderList {
		bottom = m.sortOrderList.View()
	} else {
		bottom = m.list.View()
	}

	return lipgloss.JoinVertical(lipgloss.Left, m.root.View(), bottom)
}

func sortItems(items []list.Item, field string, direction sortOrder) []list.Item {
	sortedItems := make([]list.Item, len(items))
	copy(sortedItems, items)

	switch field {
	case "last_version_date":
		switch direction {
		case sortOrderAsc:
			sort.Slice(sortedItems, func(i, j int) bool {
				a := sortedItems[i].(utils.SimpleItem)
				b := sortedItems[j].(utils.SimpleItem)
				aMod := a.Extra.(ficsit.ModsGetModsModsMod)
				bMod := b.Extra.(ficsit.ModsGetModsModsMod)
				return aMod.Last_version_date.Before(bMod.Last_version_date)
			})
		default:
			sort.Slice(sortedItems, func(i, j int) bool {
				a := sortedItems[i].(utils.SimpleItem)
				b := sortedItems[j].(utils.SimpleItem)
				aMod := a.Extra.(ficsit.ModsGetModsModsMod)
				bMod := b.Extra.(ficsit.ModsGetModsModsMod)
				return aMod.Last_version_date.After(bMod.Last_version_date)
			})
		}
	case "created_at":
		switch direction {
		case sortOrderAsc:
			sort.Slice(sortedItems, func(i, j int) bool {
				a := sortedItems[i].(utils.SimpleItem)
				b := sortedItems[j].(utils.SimpleItem)
				aMod := a.Extra.(ficsit.ModsGetModsModsMod)
				bMod := b.Extra.(ficsit.ModsGetModsModsMod)
				return aMod.Created_at.Before(bMod.Created_at)
			})
		default:
			sort.Slice(sortedItems, func(i, j int) bool {
				a := sortedItems[i].(utils.SimpleItem)
				b := sortedItems[j].(utils.SimpleItem)
				aMod := a.Extra.(ficsit.ModsGetModsModsMod)
				bMod := b.Extra.(ficsit.ModsGetModsModsMod)
				return aMod.Created_at.After(bMod.Created_at)
			})
		}
	case "name":
		switch direction {
		case sortOrderAsc:
			sort.Slice(sortedItems, func(i, j int) bool {
				a := sortedItems[i].(utils.SimpleItem)
				b := sortedItems[j].(utils.SimpleItem)
				aMod := a.Extra.(ficsit.ModsGetModsModsMod)
				bMod := b.Extra.(ficsit.ModsGetModsModsMod)
				return aMod.Name < bMod.Name
			})
		default:
			sort.Slice(sortedItems, func(i, j int) bool {
				a := sortedItems[i].(utils.SimpleItem)
				b := sortedItems[j].(utils.SimpleItem)
				aMod := a.Extra.(ficsit.ModsGetModsModsMod)
				bMod := b.Extra.(ficsit.ModsGetModsModsMod)
				return aMod.Name > bMod.Name
			})
		}
	}

	return sortedItems
}

func (m modsList) processActivation(item utils.SimpleItem, msg tea.Msg) (tea.Model, tea.Cmd) {
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
