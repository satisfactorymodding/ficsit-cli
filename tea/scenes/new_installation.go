package scenes

import (
	"os"
	"path"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/sahilm/fuzzy"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/satisfactorymodding/ficsit-cli/tea/components"
	"github.com/satisfactorymodding/ficsit-cli/tea/utils"
)

var _ tea.Model = (*newInstallation)(nil)

type newInstallation struct {
	root    components.RootModel
	parent  tea.Model
	input   textinput.Model
	title   string
	error   *components.ErrorComponent
	dirList list.Model
}

func NewNewInstallation(root components.RootModel, parent tea.Model) tea.Model {
	l := list.New([]list.Item{}, utils.NewItemDelegate(), root.Size().Width, root.Size().Height-root.Height())
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(false)
	l.SetSpinner(spinner.MiniDot)
	l.SetShowTitle(false)
	l.Styles = utils.ListStyles
	l.SetSize(l.Width(), l.Height())
	l.KeyMap.Quit.SetHelp("esc", "back")
	l.DisableQuitKeybindings()
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)

	model := newInstallation{
		root:    root,
		parent:  parent,
		input:   textinput.New(),
		title:   utils.NonListTitleStyle.Render("New Installation"),
		dirList: l,
	}

	model.input.Focus()
	model.input.Width = root.Size().Width

	// TODO SSH/FTP/SFTP support

	return model
}

func (m newInstallation) Init() tea.Cmd {
	return nil
}

func (m newInstallation) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case KeyControlC:
			return m, tea.Quit
		case KeyEscape:
			return m.parent, nil
		case KeyEnter:
			newInstall, err := m.root.GetGlobal().Installations.AddInstallation(m.root.GetGlobal(), m.input.Value(), m.root.GetGlobal().Profiles.SelectedProfile)
			if err != nil {
				errorComponent, cmd := components.NewErrorComponent(err.Error(), time.Second*5)
				m.error = errorComponent
				return m, cmd
			}

			if m.root.GetCurrentInstallation() == nil {
				if err := m.root.SetCurrentInstallation(newInstall); err != nil {
					errorComponent, cmd := components.NewErrorComponent(err.Error(), time.Second*5)
					m.error = errorComponent
					return m, cmd
				}
			}

			return m.parent, updateInstallationListCmd
		case KeyTab:
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)

			newDirItem := m.dirList.SelectedItem()

			if newDirItem == nil {
				break
			}

			newDir := newDirItem.(utils.SimpleItem[newInstallation]).ItemTitle

			newPath := ""
			_, err := os.ReadDir(m.input.Value())
			if err == nil {
				newPath = path.Join(m.input.Value(), newDir)
			} else {
				newPath = path.Join(path.Dir(m.input.Value()), newDir)
			}

			m.input.SetValue(newPath + string(os.PathSeparator))
			m.input.CursorEnd()

			listCmd := m.dirList.SetItems(getDirItems(newPath))
			m.dirList.ResetSelected()

			return m, tea.Batch(cmd, listCmd)
		default:
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)

			cmd = tea.Batch(cmd, m.dirList.SetItems(getDirItems(m.input.Value())))

			if m.dirList.Index() > len(m.dirList.Items())-1 {
				m.dirList.ResetSelected()
			}

			if key.Matches(msg, m.dirList.KeyMap.CursorUp) || key.Matches(msg, m.dirList.KeyMap.CursorDown) {
				var dirCmd tea.Cmd
				m.dirList, dirCmd = m.dirList.Update(msg)
				cmd = tea.Batch(cmd, dirCmd)
			}

			return m, cmd
		}
	case tea.WindowSizeMsg:
		m.root.SetSize(msg)
	case components.ErrorComponentTimeoutMsg:
		m.error = nil
	}

	return m, nil
}

func (m newInstallation) View() string {
	inputView := lipgloss.NewStyle().Padding(1, 2).Render(m.input.View())

	if m.error != nil {
		return lipgloss.JoinVertical(lipgloss.Left, m.root.View(), m.title, (*m.error).View(), inputView)
	}

	mandatory := lipgloss.JoinVertical(lipgloss.Left, m.root.View(), m.title, inputView)

	if len(m.dirList.Items()) > 0 {
		m.dirList.SetSize(m.dirList.Width(), m.root.Size().Height-lipgloss.Height(mandatory))

		return lipgloss.JoinVertical(lipgloss.Left, mandatory, m.dirList.View())
	}

	return mandatory
}

func getDirItems(inputValue string) []list.Item {
	filter := ""
	dir, err := os.ReadDir(inputValue)
	if err != nil {
		dir, err = os.ReadDir(path.Dir(inputValue))
		if err == nil {
			filter = path.Base(inputValue)
		}
	}

	newItems := make([]list.Item, 0)

	if inputValue != "" {
		for _, entry := range dir {
			if entry.IsDir() {
				if filter != "" {
					matches := fuzzy.Find(filter, []string{entry.Name()})
					if len(matches) == 0 {
						continue
					}
				}

				newItems = append(newItems, utils.SimpleItem[newInstallation]{
					ItemTitle: entry.Name(),
				})
			}
		}
	}

	return newItems
}
