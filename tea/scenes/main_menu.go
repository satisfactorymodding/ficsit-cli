package scenes

import (
	"fmt"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"io"
)

type mainMenu struct {
	root       RootModel
	help       help.Model
	inputStyle lipgloss.Style
	lastKey    string
	quitting   bool
	list       list.Model
}

type menuItem struct {
	Title   string
	ModelFn func(model RootModel) tea.Model
}

func (i menuItem) FilterValue() string { return i.Title }

type itemDelegate struct{}

func (d itemDelegate) Height() int                               { return 1 }
func (d itemDelegate) Spacing() int                              { return 0 }
func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(menuItem)
	if !ok {
		return
	}

	style := lipgloss.NewStyle().PaddingLeft(2)

	str := style.Render("o " + i.Title)
	if index == m.Index() {
		str = style.Foreground(lipgloss.Color("202")).Render("• " + i.Title)
	}

	fmt.Fprintf(w, str)
}

func NewMainMenu(root RootModel) tea.Model {
	items := []list.Item{
		menuItem{
			Title:   "Installations",
			ModelFn: NewInstallations,
		},
		menuItem{
			Title:   "Profiles",
			ModelFn: NewProfiles,
		},
		menuItem{
			Title:   "Mods",
			ModelFn: NewMods,
		},
	}

	l := list.NewModel(items, itemDelegate{}, 20, 14)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowTitle(false)
	l.Styles.PaginationStyle = list.DefaultStyles().PaginationStyle.PaddingLeft(2)
	l.Styles.HelpStyle = list.DefaultStyles().HelpStyle.PaddingLeft(2).PaddingBottom(1)

	return mainMenu{
		root: root,
		list: l,
	}
}

func (m mainMenu) Init() tea.Cmd {
	return nil
}

func (m mainMenu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c":
			fallthrough
		case "q":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			i, ok := m.list.SelectedItem().(menuItem)
			if ok {
				if i.ModelFn != nil {
					m.root.ChangeScene(i.ModelFn(m.root))
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
		top, right, bottom, left := lipgloss.NewStyle().Margin(2, 2).GetMargin()
		m.list.SetSize(msg.Width-left-right, msg.Height-top-bottom)
	}

	return m, nil
}

func (m mainMenu) View() string {
	return m.list.View()
}
