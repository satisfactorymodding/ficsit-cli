package scenes

import (
	"context"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/PuerkitoBio/goquery"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/satisfactorymodding/ficsit-cli/ficsit"
	"github.com/satisfactorymodding/ficsit-cli/tea/components"
	"github.com/satisfactorymodding/ficsit-cli/tea/utils"
)

var _ tea.Model = (*modVersionMenu)(nil)

type modInfo struct {
	root     components.RootModel
	viewport viewport.Model
	spinner  spinner.Model
	parent   tea.Model
	modData  chan ficsit.GetModGetMod
	ready    bool
	help     help.Model
	keys     modInfoKeyMap
}

type modInfoKeyMap struct {
	Up       key.Binding
	UpHalf   key.Binding
	UpPage   key.Binding
	Down     key.Binding
	DownHalf key.Binding
	DownPage key.Binding
	Help     key.Binding
	Back     key.Binding
}

func (k modInfoKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Back}
}

func (k modInfoKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.UpHalf, k.UpPage},
		{k.Down, k.DownHalf, k.DownPage},
		{k.Help, k.Back},
	}
}

func NewModInfo(root components.RootModel, parent tea.Model, mod utils.Mod) tea.Model {
	model := modInfo{
		root:     root,
		viewport: viewport.Model{},
		spinner:  spinner.New(),
		parent:   parent,
		modData:  make(chan ficsit.GetModGetMod),
		ready:    false,
		help:     help.New(),
		keys: modInfoKeyMap{
			Up:       key.NewBinding(key.WithHelp("↑/k", "move up")),
			UpHalf:   key.NewBinding(key.WithHelp("u", "up half page")),
			UpPage:   key.NewBinding(key.WithHelp("pgup/b", "page up")),
			Down:     key.NewBinding(key.WithHelp("↓/j", "move down")),
			DownHalf: key.NewBinding(key.WithHelp("d", "down half page")),
			DownPage: key.NewBinding(key.WithHelp("pgdn/ /f", "page down")),
			Help:     key.NewBinding(key.WithHelp("?", "toggle help")),
			Back:     key.NewBinding(key.WithHelp("q", "back")),
		},
	}

	model.spinner.Spinner = spinner.MiniDot
	model.help.Width = root.Size().Width

	go func() {
		fullMod, err := ficsit.GetMod(context.TODO(), root.GetAPIClient(), mod.ID)

		if err != nil {
			panic(err) // TODO Handle Error
		}

		if fullMod == nil {
			panic("mod is nil") // TODO Handle Error
		}

		model.modData <- fullMod.GetMod
	}()

	return model
}

func (m modInfo) Init() tea.Cmd {
	return tea.Batch(utils.Ticker(), spinner.Tick)
}

func (m modInfo) CalculateSizes(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	if m.viewport.Width == 0 {
		return m, nil
	}

	bottomPadding := 2
	if m.help.ShowAll {
		bottomPadding = 4
	}

	top, right, bottom, left := lipgloss.NewStyle().Margin(m.root.Height(), 3, bottomPadding).GetMargin()
	m.viewport.Width = msg.Width - left - right
	m.viewport.Height = msg.Height - top - bottom
	m.root.SetSize(msg)

	m.help.Width = m.viewport.Width

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m modInfo) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		case "?":
			m.help.ShowAll = !m.help.ShowAll
			newModel, cmd := m.CalculateSizes(m.root.Size())
			return newModel, cmd
		default:
			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		}
	case tea.WindowSizeMsg:
		return m.CalculateSizes(msg)
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case utils.TickMsg:
		select {
		case mod := <-m.modData:
			bottomPadding := 2
			if m.help.ShowAll {
				bottomPadding = 4
			}

			top, right, bottom, left := lipgloss.NewStyle().Margin(m.root.Height(), 3, bottomPadding).GetMargin()
			m.viewport = viewport.Model{Width: m.root.Size().Width - left - right, Height: m.root.Size().Height - top - bottom}

			title := lipgloss.NewStyle().Padding(0, 2).Render(utils.TitleStyle.Render(mod.Name)) + "\n"

			sidebar := ""
			sidebar += utils.LabelStyle.Render("Views: ") + strconv.Itoa(mod.Views) + "\n"
			sidebar += utils.LabelStyle.Render("Downloads: ") + strconv.Itoa(mod.Downloads) + "\n"
			sidebar += "\n"
			sidebar += utils.LabelStyle.Render("Authors:") + "\n"

			for _, author := range mod.Authors {
				sidebar += "\n"
				sidebar += utils.LabelStyle.Render(author.User.Username) + " - " + author.Role
			}

			converter := md.NewConverter("", true, nil)
			converter.AddRules(md.Rule{
				Filter: []string{"#text"},
				Replacement: func(content string, selec *goquery.Selection, options *md.Options) *string {
					text := selec.Text()
					return &text
				},
			})

			markdownDescription, err := converter.ConvertString(mod.Full_description)
			if err != nil {
				log.Error().Err(err).Msg("failed to convert html to markdown")
				markdownDescription = mod.Full_description
			}

			description, err := glamour.Render(markdownDescription, "dark")
			if err != nil {
				log.Error().Err(err).Msg("failed to render markdown")
				description = mod.Full_description
			}

			bottomPart := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, strings.TrimSpace(description))

			m.viewport.SetContent(lipgloss.JoinVertical(lipgloss.Left, title, bottomPart))

			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		default:
			return m, utils.Ticker()
		}
	}

	return m, nil
}

func (m modInfo) View() string {
	if m.viewport.Height == 0 {
		spinnerView := lipgloss.NewStyle().Padding(0, 2, 1).Render(m.spinner.View() + " Loading...")
		return lipgloss.JoinVertical(lipgloss.Left, m.root.View(), spinnerView)
	}

	helpBar := lipgloss.NewStyle().Padding(1, 2).Render(m.help.View(m.keys))
	return lipgloss.JoinVertical(lipgloss.Left, m.root.View(), m.viewport.View(), helpBar)
}
