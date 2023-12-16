package mods

import (
	"context"
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/truncate"

	"github.com/satisfactorymodding/ficsit-cli/cli"
	"github.com/satisfactorymodding/ficsit-cli/ficsit"
	"github.com/satisfactorymodding/ficsit-cli/tea/components"
	"github.com/satisfactorymodding/ficsit-cli/tea/scenes/keys"
	"github.com/satisfactorymodding/ficsit-cli/tea/utils"
)

var _ tea.Model = (*modsList)(nil)

type (
	sortOrder string
	sortField string
)

const (
	sortOrderAsc  sortOrder = "asc"
	sortOrderDesc sortOrder = "desc"

	sortFieldCreatedAt       sortField = "created_at"
	sortFieldDownloads       sortField = "downloads"
	sortFieldHotness         sortField = "hotness"
	sortFieldLastVersionDate sortField = "last_version_date"
	sortFieldName            sortField = "name"
	sortFieldPopularity      sortField = "popularity"
	sortFieldViews           sortField = "views"

	DefaultModSortingOrder sortOrder = sortOrderAsc
	DefaultModSortingField sortField = sortFieldLastVersionDate
)

const modsTitle = "Mods"

type listUpdate struct {
	Items []list.Item
	Done  bool
}

// type keys

type modsList struct {
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

func NewMods(root components.RootModel, parent tea.Model) tea.Model {
	l := list.New([]list.Item{}, ListDelegate{
		ItemDelegate: utils.NewItemDelegate(),
		Context:      root.GetGlobal(),
	}, root.Size().Width, root.Size().Height-root.Height())
	l.SetShowStatusBar(true)
	l.SetShowFilter(true)
	l.SetFilteringEnabled(true)
	l.SetSpinner(spinner.MiniDot)
	l.Title = modsTitle
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

	m := &modsList{
		root:         root,
		list:         l,
		parent:       parent,
		items:        make(chan listUpdate),
		sortingField: DefaultModSortingField,
		sortingOrder: DefaultModSortingOrder,
		err:          make(chan string),
	}

	m.sortFieldList = m.newSortFieldsList(root)
	m.sortOrderList = m.newSortOrderList(root)

	go func() {
		items := make([]list.Item, 0)
		allMods := make([]ficsit.ModsModsGetModsModsMod, 0)
		offset := 0
		for {
			mods, err := root.GetProvider().Mods(context.TODO(), ficsit.ModFilter{
				Limit:    100,
				Offset:   offset,
				Order_by: ficsit.ModFieldsLastVersionDate,
				Order:    ficsit.OrderDesc,
			})
			if err != nil {
				m.err <- err.Error()
				return
			}

			if len(mods.Mods.Mods) == 0 {
				break
			}

			allMods = append(allMods, mods.Mods.Mods...)

			for i := 0; i < len(mods.Mods.Mods); i++ {
				currentOffset := offset
				currentI := i
				items = append(items, utils.SimpleItemExtra[modsList, ficsit.ModsModsGetModsModsMod]{
					SimpleItem: utils.SimpleItem[modsList]{
						ItemTitle: mods.Mods.Mods[i].Name,
						Activate: func(msg tea.Msg, currentModel modsList) (tea.Model, tea.Cmd) {
							mod := allMods[currentOffset+currentI]
							return NewModMenu(root, currentModel, utils.Mod{
								Name:      mod.Name,
								Reference: mod.Mod_reference,
							}), nil
						},
					},
					Extra: allMods[currentOffset+currentI],
				})
			}

			offset += len(mods.Mods.Mods)

			m.items <- listUpdate{
				Items: items,
				Done:  false,
			}
		}

		m.items <- listUpdate{
			Items: items,
			Done:  true,
		}
	}()

	return m
}

func (m modsList) Init() tea.Cmd {
	if len(m.list.Items()) > 0 {
		return nil
	}

	return utils.Ticker()
}

func (m modsList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.list.KeyMap.Quit.SetHelp("q", "back")
	m.list.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
		}
	}

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
				i, ok := m.sortFieldList.SelectedItem().(utils.SimpleItemExtra[modsList, ficsit.ModsModsGetModsModsMod])
				if ok {
					return m.processActivation(i, msg)
				}
				return m, nil
			}

			if m.showSortOrderList {
				m.showSortOrderList = false
				i, ok := m.sortOrderList.SelectedItem().(utils.SimpleItemExtra[modsList, ficsit.ModsModsGetModsModsMod])
				if ok {
					return m.processActivation(i, msg)
				}
				return m, nil
			}

			i, ok := m.list.SelectedItem().(utils.SimpleItemExtra[modsList, ficsit.ModsModsGetModsModsMod])
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

func (m modsList) View() string {
	var bottomList list.Model
	if m.showSortFieldList {
		bottomList = m.sortFieldList
	} else if m.showSortOrderList {
		bottomList = m.sortOrderList
	} else {
		bottomList = m.list
	}

	if m.error != nil {
		err := m.error.View()
		bottomList.SetSize(bottomList.Width(), m.root.Size().Height-m.root.Height()-lipgloss.Height(err))
		return lipgloss.JoinVertical(lipgloss.Left, m.root.View(), err, bottomList.View())
	}

	bottomList.SetSize(bottomList.Width(), m.root.Size().Height-m.root.Height())
	return lipgloss.JoinVertical(lipgloss.Left, m.root.View(), bottomList.View())
}

func (m modsList) sortItems(items []list.Item, field sortField, direction sortOrder) []list.Item {
	sortedItems := make([]list.Item, len(items))
	copy(sortedItems, items)

	switch field {
	case sortFieldLastVersionDate:
		sort.Slice(sortedItems, func(i, j int) bool {
			a := sortedItems[i].(utils.SimpleItemExtra[modsList, ficsit.ModsModsGetModsModsMod])
			b := sortedItems[j].(utils.SimpleItemExtra[modsList, ficsit.ModsModsGetModsModsMod])
			return ascDesc(direction, a.Extra.Last_version_date.Before(b.Extra.Last_version_date))
		})
	case sortFieldCreatedAt:
		sort.Slice(sortedItems, func(i, j int) bool {
			a := sortedItems[i].(utils.SimpleItemExtra[modsList, ficsit.ModsModsGetModsModsMod])
			b := sortedItems[j].(utils.SimpleItemExtra[modsList, ficsit.ModsModsGetModsModsMod])
			return ascDesc(direction, a.Extra.Created_at.Before(b.Extra.Created_at))
		})
	case sortFieldName:
		sort.Slice(sortedItems, func(i, j int) bool {
			a := sortedItems[i].(utils.SimpleItemExtra[modsList, ficsit.ModsModsGetModsModsMod])
			b := sortedItems[j].(utils.SimpleItemExtra[modsList, ficsit.ModsModsGetModsModsMod])
			return ascDesc(direction, a.Extra.Name < b.Extra.Name)
		})
	case sortFieldDownloads:
		sort.Slice(sortedItems, func(i, j int) bool {
			a := sortedItems[i].(utils.SimpleItemExtra[modsList, ficsit.ModsModsGetModsModsMod])
			b := sortedItems[j].(utils.SimpleItemExtra[modsList, ficsit.ModsModsGetModsModsMod])
			return ascDesc(direction, a.Extra.Downloads < b.Extra.Downloads)
		})
	case sortFieldViews:
		sort.Slice(sortedItems, func(i, j int) bool {
			a := sortedItems[i].(utils.SimpleItemExtra[modsList, ficsit.ModsModsGetModsModsMod])
			b := sortedItems[j].(utils.SimpleItemExtra[modsList, ficsit.ModsModsGetModsModsMod])
			return ascDesc(direction, a.Extra.Views < b.Extra.Views)
		})
	case sortFieldPopularity:
		sort.Slice(sortedItems, func(i, j int) bool {
			a := sortedItems[i].(utils.SimpleItemExtra[modsList, ficsit.ModsModsGetModsModsMod])
			b := sortedItems[j].(utils.SimpleItemExtra[modsList, ficsit.ModsModsGetModsModsMod])
			return ascDesc(direction, a.Extra.Popularity < b.Extra.Popularity)
		})
	case sortFieldHotness:
		sort.Slice(sortedItems, func(i, j int) bool {
			a := sortedItems[i].(utils.SimpleItemExtra[modsList, ficsit.ModsModsGetModsModsMod])
			b := sortedItems[j].(utils.SimpleItemExtra[modsList, ficsit.ModsModsGetModsModsMod])
			return ascDesc(direction, a.Extra.Hotness < b.Extra.Hotness)
		})
	}

	return sortedItems
}

func (m modsList) processActivation(item utils.SimpleItemExtra[modsList, ficsit.ModsModsGetModsModsMod], msg tea.Msg) (tea.Model, tea.Cmd) {
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

func ascDesc(order sortOrder, result bool) bool {
	if order == sortOrderAsc {
		return result
	}
	return !result
}

type ListDelegate struct {
	list.ItemDelegate
	Context *cli.GlobalContext
}

func (c ListDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	realItem := item.(utils.SimpleItemExtra[modsList, ficsit.ModsModsGetModsModsMod])
	realDelegate := c.ItemDelegate.(list.DefaultDelegate)

	title := realItem.Title()

	s := &realDelegate.Styles

	if m.Width() <= 0 {
		return
	}

	textwidth := uint(m.Width() - s.NormalTitle.GetPaddingLeft() - s.NormalTitle.GetPaddingRight())
	title = truncate.StringWithTail(title, textwidth, "…")

	var (
		isSelected  = index == m.Index()
		emptyFilter = m.FilterState() == list.Filtering && m.FilterValue() == ""
		isFiltered  = m.FilterState() == list.Filtering || m.FilterState() == list.FilterApplied
	)

	var matchedRunes []int
	if isFiltered && index < len(m.VisibleItems()) {
		// Get indices of matched characters
		matchedRunes = m.MatchesForItem(index)
	}

	isInstalled := false
	isDisabled := false
	if c.Context != nil {
		profile := c.Context.Profiles.Profiles[c.Context.Profiles.SelectedProfile]
		if profile != nil {
			if profile.HasMod(realItem.Extra.Mod_reference) {
				isInstalled = true
				isDisabled = !profile.IsModEnabled(realItem.Extra.Mod_reference)
			}
		}
	}

	if emptyFilter {
		if isInstalled {
			if isDisabled {
				title = lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Render("✓ " + title)
			} else {
				title = lipgloss.NewStyle().Foreground(lipgloss.Color("40")).Render("✓ " + title)
			}
		}
		title = s.DimmedTitle.Render(title)
	} else if isSelected && m.FilterState() != list.Filtering {
		if isFiltered {
			unmatched := s.SelectedTitle.Inline(true)
			matched := unmatched.Copy().Inherit(s.FilterMatch)
			if isInstalled {
				if isDisabled {
					unmatched = unmatched.Foreground(lipgloss.Color("220"))
					matched = matched.Foreground(lipgloss.Color("220"))
				} else {
					unmatched = unmatched.Foreground(lipgloss.Color("40"))
					matched = matched.Foreground(lipgloss.Color("40"))
				}
			}
			title = lipgloss.StyleRunes(title, matchedRunes, matched, unmatched)
		}
		if isInstalled {
			if isDisabled {
				title = lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Render("✓ ") + title
			} else {
				title = lipgloss.NewStyle().Foreground(lipgloss.Color("40")).Render("✓ ") + title
			}
		}
		title = s.SelectedTitle.Render(title)
	} else {
		if isFiltered {
			unmatched := s.NormalTitle.Inline(true)
			matched := unmatched.Copy().Inherit(s.FilterMatch)
			if isInstalled {
				if isDisabled {
					unmatched = unmatched.Foreground(lipgloss.Color("220"))
					matched = matched.Foreground(lipgloss.Color("220"))
				} else {
					unmatched = unmatched.Foreground(lipgloss.Color("40"))
					matched = matched.Foreground(lipgloss.Color("40"))
				}
			}
			title = lipgloss.StyleRunes(title, matchedRunes, matched, unmatched)
		}
		if isInstalled {
			if isDisabled {
				title = lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Render("✓ ") + title
			} else {
				title = lipgloss.NewStyle().Foreground(lipgloss.Color("40")).Render("✓ ") + title
			}
		}
		title = s.NormalTitle.Render(title)
	}

	fmt.Fprintf(w, "%s", title)
}

func (m modsList) newSortFieldsList(root components.RootModel) list.Model {
	sortFieldList := list.New([]list.Item{
		utils.SimpleItemExtra[modsList, ficsit.ModsModsGetModsModsMod]{
			SimpleItem: utils.SimpleItem[modsList]{
				ItemTitle: "Name",
				Activate: func(msg tea.Msg, m modsList) (tea.Model, tea.Cmd) {
					m.sortingField = sortFieldName
					cmd := m.list.SetItems(m.sortItems(m.list.Items(), m.sortingField, m.sortingOrder))
					m.list.ResetSelected()
					return m, cmd
				},
			},
		},
		utils.SimpleItemExtra[modsList, ficsit.ModsModsGetModsModsMod]{
			SimpleItem: utils.SimpleItem[modsList]{
				ItemTitle: "Last Version Date",
				Activate: func(msg tea.Msg, m modsList) (tea.Model, tea.Cmd) {
					m.sortingField = sortFieldLastVersionDate
					cmd := m.list.SetItems(m.sortItems(m.list.Items(), m.sortingField, m.sortingOrder))
					m.list.ResetSelected()
					return m, cmd
				},
			},
		},
		utils.SimpleItemExtra[modsList, ficsit.ModsModsGetModsModsMod]{
			SimpleItem: utils.SimpleItem[modsList]{
				ItemTitle: "Creation Date",
				Activate: func(msg tea.Msg, m modsList) (tea.Model, tea.Cmd) {
					m.sortingField = sortFieldCreatedAt
					cmd := m.list.SetItems(m.sortItems(m.list.Items(), m.sortingField, m.sortingOrder))
					m.list.ResetSelected()
					return m, cmd
				},
			},
		},
		utils.SimpleItemExtra[modsList, ficsit.ModsModsGetModsModsMod]{
			SimpleItem: utils.SimpleItem[modsList]{
				ItemTitle: "Downloads",
				Activate: func(msg tea.Msg, m modsList) (tea.Model, tea.Cmd) {
					m.sortingField = sortFieldDownloads
					cmd := m.list.SetItems(m.sortItems(m.list.Items(), m.sortingField, m.sortingOrder))
					m.list.ResetSelected()
					return m, cmd
				},
			},
		},
		utils.SimpleItemExtra[modsList, ficsit.ModsModsGetModsModsMod]{
			SimpleItem: utils.SimpleItem[modsList]{
				ItemTitle: "Views",
				Activate: func(msg tea.Msg, m modsList) (tea.Model, tea.Cmd) {
					m.sortingField = sortFieldViews
					cmd := m.list.SetItems(m.sortItems(m.list.Items(), m.sortingField, m.sortingOrder))
					m.list.ResetSelected()
					return m, cmd
				},
			},
		},
		utils.SimpleItemExtra[modsList, ficsit.ModsModsGetModsModsMod]{
			SimpleItem: utils.SimpleItem[modsList]{
				ItemTitle: "Popularity (recent downloads)",
				Activate: func(msg tea.Msg, m modsList) (tea.Model, tea.Cmd) {
					m.sortingField = sortFieldPopularity
					cmd := m.list.SetItems(m.sortItems(m.list.Items(), m.sortingField, m.sortingOrder))
					m.list.ResetSelected()
					return m, cmd
				},
			},
		},
		utils.SimpleItemExtra[modsList, ficsit.ModsModsGetModsModsMod]{
			SimpleItem: utils.SimpleItem[modsList]{
				ItemTitle: "Hotness (recent views)",
				Activate: func(msg tea.Msg, m modsList) (tea.Model, tea.Cmd) {
					m.sortingField = sortFieldHotness
					cmd := m.list.SetItems(m.sortItems(m.list.Items(), m.sortingField, m.sortingOrder))
					m.list.ResetSelected()
					return m, cmd
				},
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

	return sortFieldList
}

func (m modsList) newSortOrderList(root components.RootModel) list.Model {
	sortOrderList := list.New([]list.Item{
		utils.SimpleItemExtra[modsList, ficsit.ModsModsGetModsModsMod]{
			SimpleItem: utils.SimpleItem[modsList]{
				ItemTitle: "Ascending",
				Activate: func(msg tea.Msg, m modsList) (tea.Model, tea.Cmd) {
					m.sortingOrder = sortOrderAsc
					cmd := m.list.SetItems(m.sortItems(m.list.Items(), m.sortingField, m.sortingOrder))
					m.list.ResetSelected()
					return m, cmd
				},
			},
		},
		utils.SimpleItemExtra[modsList, ficsit.ModsModsGetModsModsMod]{
			SimpleItem: utils.SimpleItem[modsList]{
				ItemTitle: "Descending",
				Activate: func(msg tea.Msg, m modsList) (tea.Model, tea.Cmd) {
					m.sortingOrder = sortOrderDesc
					cmd := m.list.SetItems(m.sortItems(m.list.Items(), m.sortingField, m.sortingOrder))
					m.list.ResetSelected()
					return m, cmd
				},
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

	return sortOrderList
}
