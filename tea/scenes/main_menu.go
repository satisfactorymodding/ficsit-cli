package scenes

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rs/zerolog/log"
	"github.com/satisfactorymodding/ficsit-cli/tea/components"
	"github.com/satisfactorymodding/ficsit-cli/tea/utils"
	"github.com/spf13/viper"
)

var _ tea.Model = (*mainMenu)(nil)

type mainMenu struct {
	root   components.RootModel
	list   list.Model
	error  *components.ErrorComponent
	banner string
}

const logoBanner = `
███████╗██╗ ██████╗███████╗██╗████████╗
██╔════╝██║██╔════╝██╔════╝██║╚══██╔══╝
█████╗  ██║██║     ███████╗██║   ██║
██╔══╝  ██║██║     ╚════██║██║   ██║
██║     ██║╚██████╗███████║██║   ██║
╚═╝     ╚═╝ ╚═════╝╚══════╝╚═╝   ╚═╝`

func NewMainMenu(root components.RootModel) tea.Model {
	trimmedBanner := strings.TrimSpace(logoBanner)
	var finalBanner strings.Builder

	for i, s := range strings.Split(trimmedBanner, "\n") {
		if i > 0 {
			finalBanner.WriteRune('\n')
		}

		foreground := utils.LogoForegroundStyles[i]
		background := utils.LogoBackgroundStyles[i]

		for _, c := range s {
			if c == '█' {
				finalBanner.WriteString(foreground.Render("█"))
			} else if c != ' ' {
				finalBanner.WriteString(background.Render(string(c)))
			} else {
				finalBanner.WriteRune(c)
			}
		}
	}

	model := mainMenu{
		root:   root,
		banner: finalBanner.String(),
	}

	items := []list.Item{
		utils.SimpleItem[mainMenu]{
			ItemTitle: "Installations",
			Activate: func(msg tea.Msg, currentModel mainMenu) (tea.Model, tea.Cmd) {
				newModel := NewInstallations(root, currentModel)
				return newModel, newModel.Init()
			},
		},
		utils.SimpleItem[mainMenu]{
			ItemTitle: "Profiles",
			Activate: func(msg tea.Msg, currentModel mainMenu) (tea.Model, tea.Cmd) {
				newModel := NewProfiles(root, currentModel)
				return newModel, newModel.Init()
			},
		},
		utils.SimpleItem[mainMenu]{
			ItemTitle: "All Mods",
			Activate: func(msg tea.Msg, currentModel mainMenu) (tea.Model, tea.Cmd) {
				newModel := NewMods(root, currentModel)
				return newModel, newModel.Init()
			},
		},
		utils.SimpleItem[mainMenu]{
			ItemTitle: "Installed Mods",
			Activate: func(msg tea.Msg, currentModel mainMenu) (tea.Model, tea.Cmd) {
				newModel := NewInstalledMods(root, currentModel)
				return newModel, newModel.Init()
			},
		},
		utils.SimpleItem[mainMenu]{
			ItemTitle: "Apply Changes",
			Activate: func(msg tea.Msg, currentModel mainMenu) (tea.Model, tea.Cmd) {
				if err := root.GetGlobal().Save(); err != nil {
					log.Error().Err(err).Msg(ErrorFailedAddMod)
					errorComponent, cmd := components.NewErrorComponent(err.Error(), time.Second*5)
					currentModel.error = errorComponent
					return currentModel, cmd
				}

				newModel := NewApply(root, currentModel)
				return newModel, newModel.Init()
			},
		},
		utils.SimpleItem[mainMenu]{
			ItemTitle: "Save",
			Activate: func(msg tea.Msg, currentModel mainMenu) (tea.Model, tea.Cmd) {
				if err := root.GetGlobal().Save(); err != nil {
					log.Error().Err(err).Msg(ErrorFailedAddMod)
					errorComponent, cmd := components.NewErrorComponent(err.Error(), time.Second*5)
					currentModel.error = errorComponent
					return currentModel, cmd
				}
				return nil, nil
			},
		},
		utils.SimpleItem[mainMenu]{
			ItemTitle: "Exit",
			Activate: func(msg tea.Msg, currentModel mainMenu) (tea.Model, tea.Cmd) {
				return nil, tea.Quit
			},
		},
	}

	model.list = list.New(items, utils.NewItemDelegate(), root.Size().Width, root.Size().Height-root.Height())
	model.list.SetShowStatusBar(false)
	model.list.SetFilteringEnabled(false)
	model.list.Title = "Main Menu"
	model.list.Styles = utils.ListStyles
	model.list.DisableQuitKeybindings()

	return model
}

func (m mainMenu) Init() tea.Cmd {
	return nil
}

func (m mainMenu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case KeyControlC:
			return m, tea.Quit
		case "q":
			return m, tea.Quit
		case KeyEnter:
			i, ok := m.list.SelectedItem().(utils.SimpleItem[mainMenu])
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
		top, right, bottom, left := lipgloss.NewStyle().Margin(2, 2).GetMargin()
		m.list.SetSize(msg.Width-left-right, msg.Height-top-bottom)
		m.root.SetSize(msg)
	case components.ErrorComponentTimeoutMsg:
		m.error = nil
	}

	return m, nil
}

func (m mainMenu) View() string {
	header := m.root.View()

	banner := lipgloss.NewStyle().Margin(2, 0, 0, 2).Render(m.banner)

	commit := viper.GetString("commit")
	if len(commit) > 8 {
		commit = commit[:8]
	}

	version := "\n"
	version += utils.LabelStyle.Render("Version: ")
	version += viper.GetString("version") + " - " + commit

	header = lipgloss.JoinVertical(lipgloss.Left, version, header)

	totalHeight := lipgloss.Height(header) + len(m.list.Items()) + lipgloss.Height(banner) + 5
	if totalHeight < m.root.Size().Height {
		header = lipgloss.JoinVertical(lipgloss.Left, banner, header)
	}

	if m.error != nil {
		err := (*m.error).View()
		m.list.SetSize(m.list.Width(), m.root.Size().Height-lipgloss.Height(header)-lipgloss.Height(err))
		return lipgloss.JoinVertical(lipgloss.Left, header, err, m.list.View())
	}

	m.list.SetSize(m.list.Width(), m.root.Size().Height-lipgloss.Height(header))
	return lipgloss.JoinVertical(lipgloss.Left, header, m.list.View())
}
