package scenes

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/davecgh/go-spew/spew"
	"github.com/rs/zerolog/log"
	"github.com/satisfactorymodding/ficsit-cli/ficsit"
	"github.com/satisfactorymodding/ficsit-cli/tea/components"
	"github.com/satisfactorymodding/ficsit-cli/tea/utils"
)

var _ tea.Model = (*selectModVersionList)(nil)

type selectModVersionList struct {
	root   components.RootModel
	list   list.Model
	parent tea.Model
	items  chan []list.Item
}

func NewModVersionList(root components.RootModel, parent tea.Model, mod utils.Mod) tea.Model {
	l := list.NewModel([]list.Item{}, utils.NewItemDelegate(), root.Size().Width, root.Size().Height-root.Height())
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(false)
	l.SetSpinner(spinner.MiniDot)
	l.Title = fmt.Sprintf("Versions (%s)", mod.Name)
	l.Styles = utils.ListStyles
	l.SetSize(l.Width(), l.Height())
	l.KeyMap.Quit.SetHelp("q", "back")

	m := &selectModVersionList{
		root:   root,
		list:   l,
		parent: parent,
		items:  make(chan []list.Item),
	}

	go func() {
		items := make([]list.Item, 0)
		allVersions := make([]ficsit.ModVersionsGetModVersionsVersion, 0)
		offset := 0
		for {
			versions, err := ficsit.ModVersions(context.TODO(), root.GetAPIClient(), mod.ID, ficsit.VersionFilter{
				Limit:    100,
				Offset:   offset,
				Order:    ficsit.OrderDesc,
				Order_by: ficsit.VersionFieldsCreatedAt,
			})

			if err != nil {
				panic(err) // TODO
			}

			if len(versions.GetMod.Versions) == 0 {
				break
			}

			allVersions = append(allVersions, versions.GetMod.Versions...)

			for i := 0; i < len(versions.GetMod.Versions); i++ {
				currentOffset := offset
				currentI := i
				items = append(items, utils.SimpleItem{
					ItemTitle: versions.GetMod.Versions[i].Version,
					Activate: func(msg tea.Msg, currentModel tea.Model) (tea.Model, tea.Cmd) {
						version := allVersions[currentOffset+currentI]
						err := root.GetCurrentProfile().AddMod(mod.Reference, version.Version)
						if err != nil {
							panic(err) // TODO
						}
						return currentModel.(selectModVersionList).parent, nil
					},
				})
			}

			offset += len(versions.GetMod.Versions)
		}

		m.items <- items
	}()

	return m
}

func (m selectModVersionList) Init() tea.Cmd {
	return utils.Ticker()
}

func (m selectModVersionList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	log.Info().Msg(spew.Sdump(msg))
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case KeyControlC:
			return m, tea.Quit
		case "q":
			if m.parent != nil {
				m.parent.Update(m.root.Size())
				return m.parent, nil
			}
			return m, tea.Quit
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
		top, right, bottom, left := lipgloss.NewStyle().Margin(m.root.Height(), 2, 0).GetMargin()
		m.list.SetSize(msg.Width-left-right, msg.Height-top-bottom)
		m.root.SetSize(msg)
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
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

	return m, nil
}

func (m selectModVersionList) View() string {
	return lipgloss.JoinVertical(lipgloss.Left, m.root.View(), m.list.View())
}
