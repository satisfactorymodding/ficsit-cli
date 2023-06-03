package scenes

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
	"github.com/satisfactorymodding/ficsit-cli/tea/utils"
)

var _ tea.Model = (*updateModsList)(nil)

type updateModsList struct {
	root   components.RootModel
	list   list.Model
	parent tea.Model
	items  chan listUpdate

	err   chan string
	error *components.ErrorComponent

	selectedMods []string
}

func NewUpdateMods(root components.RootModel, parent tea.Model) tea.Model {
	if root.GetCurrentProfile() == nil {
		return parent
	}
	if root.GetCurrentInstallation() == nil {
		return parent
	}

	l := list.New([]list.Item{}, updateModsListDelegate{ItemDelegate: utils.NewItemDelegate(), selectedMods: []string{}}, root.Size().Width, root.Size().Height-root.Height())
	l.SetShowStatusBar(true)
	l.SetShowFilter(true)
	l.SetFilteringEnabled(true)
	l.SetSpinner(spinner.MiniDot)
	l.Title = "Update Mods"
	l.Styles = utils.ListStyles
	l.SetSize(l.Width(), l.Height())
	l.KeyMap.Quit.SetHelp("q", "back")
	l.DisableQuitKeybindings()

	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(key.WithHelp("q", "back")),
			key.NewBinding(key.WithHelp("space", "select")),
			key.NewBinding(key.WithHelp("enter", "confirm")),
		}
	}

	l.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(key.WithHelp("q", "back")),
			key.NewBinding(key.WithHelp("space", "select")),
			key.NewBinding(key.WithHelp("enter", "confirm")),
		}
	}

	return &updateModsList{
		root:   root,
		list:   l,
		parent: parent,
		items:  make(chan listUpdate),
		err:    make(chan string),
	}
}

func (m updateModsList) Init() tea.Cmd {
	go m.LoadModData()
	return utils.Ticker()
}

type modUpdate struct {
	Reference string
	From      string
	To        string
}

type modToggleMsg struct {
	reference string
}

func (m updateModsList) LoadModData() {
	currentInstallation := m.root.GetCurrentInstallation()
	currentProfile := m.root.GetCurrentProfile()

	currentLockfile, err := m.root.GetCurrentInstallation().LockFile(m.root.GetGlobal())
	if err != nil {
		return
	}
	if currentLockfile == nil {
		return
	}

	gameVersion, err := currentInstallation.GetGameVersion(m.root.GetGlobal())
	if err != nil {
		return
	}

	resolver := cli.NewDependencyResolver(m.root.GetAPIClient())

	updatedLockfile, err := currentProfile.Resolve(resolver, nil, gameVersion)
	if err != nil {
		return
	}

	items := make([]list.Item, 0)
	i := 0
	for reference, currentLockedMod := range currentLockfile {
		r := reference
		updatedLockedMod, ok := updatedLockfile[reference]
		if !ok {
			continue
		}
		if updatedLockedMod.Version == currentLockedMod.Version {
			continue
		}
		items = append(items, utils.SimpleItemExtra[updateModsList, modUpdate]{
			SimpleItem: utils.SimpleItem[updateModsList]{
				ItemTitle: fmt.Sprintf("%s - %s -> %s", reference, currentLockedMod.Version, updatedLockedMod.Version),
				Activate: func(msg tea.Msg, currentModel updateModsList) (tea.Model, tea.Cmd) {
					return currentModel, func() tea.Msg {
						return modToggleMsg{reference: r}
					}
				},
			},
			Extra: modUpdate{
				Reference: r,
				From:      currentLockedMod.Version,
				To:        updatedLockedMod.Version,
			},
		})
		i++
	}

	sort.Slice(items, func(i, j int) bool {
		a := items[i].(utils.SimpleItemExtra[updateModsList, modUpdate])
		b := items[j].(utils.SimpleItemExtra[updateModsList, modUpdate])
		return ascDesc(sortOrderDesc, a.ItemTitle < b.ItemTitle)
	})

	m.items <- listUpdate{
		Items: items,
		Done:  false,
	}

	m.loadModNames(items)
}

func (m updateModsList) loadModNames(items []list.Item) {
	if len(items) == 0 {
		m.items <- listUpdate{
			Items: items,
			Done:  true,
		}
		return
	}

	references := make([]string, len(items))
	i := 0
	for _, item := range items {
		references[i] = item.(utils.SimpleItemExtra[updateModsList, modUpdate]).Extra.Reference
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

	newItems := make([]list.Item, len(mods.Mods.Mods))
	for i, mod := range mods.Mods.Mods {
		// Re-reference struct
		mod := mod
		var currentModUpdate modUpdate
		for _, item := range items {
			currentModUpdate = item.(utils.SimpleItemExtra[updateModsList, modUpdate]).Extra
			if currentModUpdate.Reference == mod.Mod_reference {
				break
			}
		}
		newItems[i] = utils.SimpleItemExtra[updateModsList, modUpdate]{
			SimpleItem: utils.SimpleItem[updateModsList]{
				ItemTitle: fmt.Sprintf("%s - %s -> %s", mod.Name, currentModUpdate.From, currentModUpdate.To),
				Activate: func(msg tea.Msg, currentModel updateModsList) (tea.Model, tea.Cmd) {
					return currentModel, func() tea.Msg {
						return modToggleMsg{reference: mod.Mod_reference}
					}
				},
			},
			Extra: currentModUpdate,
		}
	}

	sort.Slice(newItems, func(i, j int) bool {
		a := newItems[i].(utils.SimpleItemExtra[updateModsList, modUpdate])
		b := newItems[j].(utils.SimpleItemExtra[updateModsList, modUpdate])
		return ascDesc(sortOrderDesc, a.Extra.Reference < b.Extra.Reference)
	})

	m.items <- listUpdate{
		Items: newItems,
		Done:  true,
	}
}

func (m updateModsList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// List enables its own keybindings when they were previously disabled
	m.list.DisableQuitKeybindings()

	cmds := make([]tea.Cmd, 0)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.list.SettingFilter() {
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}

		switch keypress := msg.String(); keypress {
		case KeyControlC:
			return m, tea.Quit
		case "q":
			if m.parent != nil {
				m.parent.Update(m.root.Size())
				return m.parent, nil
			}
			return m, tea.Quit
		case " ":
			i, ok := m.list.SelectedItem().(utils.SimpleItem[updateModsList])
			if ok {
				return m.processActivation(i, msg)
			}
			i2, ok := m.list.SelectedItem().(utils.SimpleItemExtra[updateModsList, modUpdate])
			if ok {
				return m.processActivation(i2.SimpleItem, msg)
			}
			return m, nil
		case KeyEnter:
			if len(m.selectedMods) > 0 {
				err := m.root.GetCurrentInstallation().UpdateMods(m.root.GetGlobal(), m.selectedMods)
				if err != nil {
					m.err <- err.Error()
					return m, nil
				}
			}
			if m.parent != nil {
				m.parent.Update(m.root.Size())
				return m.parent, nil
			}
			return m, tea.Quit
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
	case modToggleMsg:
		idx := -1
		for i, mod := range m.selectedMods {
			if mod == msg.reference {
				idx = i
				break
			}
		}
		if idx != -1 {
			m.selectedMods = append(m.selectedMods[:idx], m.selectedMods[idx+1:]...)
		} else {
			m.selectedMods = append(m.selectedMods, msg.reference)
		}
		cmds = append(cmds, func() tea.Msg { return selectedModsUpdateMsg{selectedMods: m.selectedMods} })
	}

	newList, listCmd := m.list.Update(msg)
	m.list = newList
	cmds = append(cmds, listCmd)

	return m, tea.Batch(cmds...)
}

func (m updateModsList) View() string {
	m.list.SetSize(m.list.Width(), m.root.Size().Height-m.root.Height())
	return lipgloss.JoinVertical(lipgloss.Left, m.root.View(), m.list.View())
}

func (m updateModsList) processActivation(item utils.SimpleItem[updateModsList], msg tea.Msg) (tea.Model, tea.Cmd) {
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

type updateModsListDelegate struct {
	list.ItemDelegate
	selectedMods []string
}

type selectedModsUpdateMsg struct {
	selectedMods []string
}

func (c updateModsListDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	if msg, ok := msg.(selectedModsUpdateMsg); ok {
		c.selectedMods = msg.selectedMods
		m.SetDelegate(c)
	}
	return nil
}

func (c updateModsListDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	realItem := item.(utils.SimpleItemExtra[updateModsList, modUpdate])
	realDelegate := c.ItemDelegate.(list.DefaultDelegate)

	title := realItem.Title()

	s := &realDelegate.Styles

	if m.Width() <= 0 {
		return
	}

	textwidth := uint(m.Width() - s.NormalTitle.GetPaddingLeft() - s.NormalTitle.GetPaddingRight())
	title = truncate.StringWithTail(title, textwidth, "…")

	isSelected := index == m.Index()

	isUpdating := false
	for _, mod := range c.selectedMods {
		if mod == realItem.Extra.Reference {
			isUpdating = true
		}
	}

	var checkbox string
	if isUpdating {
		checkbox = lipgloss.NewStyle().Foreground(lipgloss.Color("40")).Render("[✓]")
	} else {
		checkbox = lipgloss.NewStyle().Foreground(lipgloss.Color("40")).Render("[ ]")
	}

	if isSelected {
		title = s.SelectedTitle.UnsetBorderLeft().UnsetPaddingLeft().Render(title)
	} else {
		title = s.NormalTitle.UnsetPaddingLeft().Render(title)
	}

	fmt.Fprintf(w, "%s %s", checkbox, title)
}
