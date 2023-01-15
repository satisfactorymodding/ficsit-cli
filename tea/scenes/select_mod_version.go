package scenes

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

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
	err    chan string
	error  *components.ErrorComponent
}

func NewModVersionList(root components.RootModel, parent tea.Model, mod utils.Mod) tea.Model {
	l := list.New([]list.Item{}, utils.NewItemDelegate(), root.Size().Width, root.Size().Height-root.Height())
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
		err:    make(chan string),
	}

	go func() {
		items := make([]list.Item, 0)
		allVersions := make([]ficsit.ModVersionsModVersionsVersion, 0)
		offset := 0
		for {
			versions, err := ficsit.ModVersions(context.TODO(), root.GetAPIClient(), mod.Reference, ficsit.VersionFilter{
				Limit:    100,
				Offset:   offset,
				Order:    ficsit.OrderDesc,
				Order_by: ficsit.VersionFieldsCreatedAt,
			})
			if err != nil {
				m.err <- err.Error()
				return
			}

			if len(versions.Mod.Versions) == 0 {
				break
			}

			allVersions = append(allVersions, versions.Mod.Versions...)

			for i := 0; i < len(versions.Mod.Versions); i++ {
				currentOffset := offset
				currentI := i
				items = append(items, utils.SimpleItem[selectModVersionList]{
					ItemTitle: versions.Mod.Versions[i].Version,
					Activate: func(msg tea.Msg, currentModel selectModVersionList) (tea.Model, tea.Cmd) {
						version := allVersions[currentOffset+currentI]
						err := root.GetCurrentProfile().AddMod(mod.Reference, version.Version)
						if err != nil {
							errorComponent, cmd := components.NewErrorComponent(err.Error(), time.Second*5)
							currentModel.error = errorComponent
							return currentModel, cmd
						}
						return currentModel.parent, nil
					},
				})
			}

			offset += len(versions.Mod.Versions)
		}

		m.items <- items
	}()

	return m
}

func (m selectModVersionList) Init() tea.Cmd {
	return utils.Ticker()
}

func (m selectModVersionList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			i, ok := m.list.SelectedItem().(utils.SimpleItem[selectModVersionList])
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
			return m, cmd
		case err := <-m.err:
			errorComponent, cmd := components.NewErrorComponent(err, time.Second*5)
			m.error = errorComponent
			return m, cmd
		default:
			start := m.list.StartSpinner()
			return m, tea.Batch(utils.Ticker(), start)
		}
	}

	return m, nil
}

func (m selectModVersionList) View() string {
	if m.error != nil {
		err := m.error.View()
		m.list.SetSize(m.list.Width(), m.root.Size().Height-m.root.Height()-lipgloss.Height(err))
		return lipgloss.JoinVertical(lipgloss.Left, m.root.View(), err, m.list.View())
	}

	m.list.SetSize(m.list.Width(), m.root.Size().Height-m.root.Height())
	return lipgloss.JoinVertical(lipgloss.Left, m.root.View(), m.list.View())
}
