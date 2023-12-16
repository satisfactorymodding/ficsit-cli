package installation

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/truncate"
	"github.com/sahilm/fuzzy"

	"github.com/satisfactorymodding/ficsit-cli/tea/components"
	"github.com/satisfactorymodding/ficsit-cli/tea/scenes/keys"
	"github.com/satisfactorymodding/ficsit-cli/tea/utils"
)

var _ tea.Model = (*newInstallation)(nil)

type newInstallation struct {
	dirList list.Model
	root    components.RootModel
	parent  tea.Model
	error   *components.ErrorComponent
	title   string
	input   textinput.Model
}

func NewNewInstallation(root components.RootModel, parent tea.Model) tea.Model {
	listDelegate := NewInstallListDelegate{ItemDelegate: utils.NewItemDelegate()}

	l := list.New([]list.Item{}, listDelegate, root.Size().Width, root.Size().Height-root.Height())
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(false)
	l.SetSpinner(spinner.MiniDot)
	l.SetShowTitle(false)
	l.Styles = utils.ListStyles
	l.SetSize(l.Width(), l.Height())
	l.SetShowStatusBar(false)
	l.KeyMap.ShowFullHelp.Unbind()
	l.KeyMap.Quit.SetHelp("esc", "back")
	l.KeyMap.CursorDown.SetHelp("↓", "down")
	l.KeyMap.CursorUp.SetHelp("↑", "up")
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(key.WithKeys(keys.KeyTab), key.WithHelp(keys.KeyTab, "select")),
			key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "continue")),
		}
	}

	model := newInstallation{
		root:    root,
		parent:  parent,
		input:   textinput.New(),
		title:   utils.NonListTitleStyle.Render("New Installation"),
		dirList: l,
	}

	model.input.Focus()
	model.input.Width = root.Size().Width

	return model
}

func (m newInstallation) Init() tea.Cmd {
	return textinput.Blink
}

func (m newInstallation) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case keys.KeyControlC:
			return m, tea.Quit
		case keys.KeyEscape:
			return m.parent, nil
		case keys.KeyEnter:
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
		case keys.KeyTab:
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)

			newDirItem := m.dirList.SelectedItem()

			if newDirItem == nil {
				break
			}

			newDir := newDirItem.(utils.SimpleItemExtra[newInstallation, string]).ItemTitle

			newPath := ""
			_, err := os.ReadDir(m.input.Value())
			if err == nil {
				newPath = filepath.Join(m.input.Value(), newDir)
			} else {
				newPath = filepath.Join(filepath.Dir(m.input.Value()), newDir)
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
	default:
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m newInstallation) View() string {
	style := lipgloss.NewStyle().Padding(1, 2)
	inputView := style.Render(m.input.View())

	mandatory := lipgloss.JoinVertical(lipgloss.Left, m.root.View(), m.title, inputView)

	if m.error != nil {
		return lipgloss.JoinVertical(lipgloss.Left, mandatory, m.error.View())
	}

	if len(m.dirList.Items()) == 0 {
		infoBox := lipgloss.NewStyle().
			BorderStyle(lipgloss.ThickBorder()).
			BorderForeground(lipgloss.Color("39")).
			Padding(0, 1).
			Margin(0, 0, 0, 2).
			Render("Enter the path to the satisfactory installation")
		mandatory = lipgloss.JoinVertical(lipgloss.Left, mandatory, infoBox)
	}

	m.dirList.SetSize(m.dirList.Width(), m.root.Size().Height-lipgloss.Height(mandatory)-1)
	return lipgloss.JoinVertical(lipgloss.Left, mandatory, m.dirList.View())
}

// I know this is awful, but beats re-implementing the entire list model
var globalMatches []fuzzy.Match

func getDirItems(inputValue string) []list.Item {
	filter := ""
	dir, err := os.ReadDir(inputValue)
	if err != nil {
		dir, err = os.ReadDir(filepath.Dir(inputValue))
		if err == nil {
			filter = filepath.Base(inputValue)
		}
	}

	newItems := make([]list.Item, 0)

	globalMatches = nil

	if inputValue != "" {
		if filter != "" {
			dirNames := make([]string, 0)
			for _, entry := range dir {
				if entry.IsDir() || entry.Type() == fs.ModeSymlink {
					dirNames = append(dirNames, entry.Name())
				}
			}

			matches := fuzzy.Find(filter, dirNames)
			sort.Stable(matches)

			for _, match := range matches {
				newItems = append(newItems, utils.SimpleItemExtra[newInstallation, string]{
					SimpleItem: utils.SimpleItem[newInstallation]{
						ItemTitle: match.Str,
					},
					Extra: match.Str,
				})
			}

			globalMatches = matches
		} else {
			for _, entry := range dir {
				if entry.IsDir() || entry.Type() == fs.ModeSymlink {
					newItems = append(newItems, utils.SimpleItemExtra[newInstallation, string]{
						SimpleItem: utils.SimpleItem[newInstallation]{
							ItemTitle: entry.Name(),
						},
						Extra: entry.Name(),
					})
				}
			}
		}
	}

	return newItems
}

type NewInstallListDelegate struct {
	list.ItemDelegate
}

func (c NewInstallListDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	realItem := item.(utils.SimpleItemExtra[newInstallation, string])
	realDelegate := c.ItemDelegate.(list.DefaultDelegate)

	title := realItem.Title()

	s := &realDelegate.Styles

	if m.Width() <= 0 {
		return
	}

	textwidth := uint(m.Width() - s.NormalTitle.GetPaddingLeft() - s.NormalTitle.GetPaddingRight())
	title = truncate.StringWithTail(title, textwidth, "…")

	if index == m.Index() {
		if globalMatches != nil {
			unmatched := s.SelectedTitle.Inline(true)
			matched := unmatched.Copy().Inherit(s.FilterMatch)
			title = lipgloss.StyleRunes(title, globalMatches[index].MatchedIndexes, matched, unmatched)
		}
		title = s.SelectedTitle.Render(title)
	} else {
		if globalMatches != nil {
			unmatched := s.NormalTitle.Inline(true)
			matched := unmatched.Copy().Inherit(s.FilterMatch)
			title = lipgloss.StyleRunes(title, globalMatches[index].MatchedIndexes, matched, unmatched)
		}
		title = s.NormalTitle.Render(title)
	}

	fmt.Fprintf(w, "%s", title)
}
