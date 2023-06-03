package scenes

import (
	"sort"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wrap"

	"github.com/satisfactorymodding/ficsit-cli/cli"
	"github.com/satisfactorymodding/ficsit-cli/tea/components"
	"github.com/satisfactorymodding/ficsit-cli/tea/scenes/keys"
	teaUtils "github.com/satisfactorymodding/ficsit-cli/tea/utils"
	"github.com/satisfactorymodding/ficsit-cli/utils"
)

var _ tea.Model = (*apply)(nil)

type modProgress struct {
	downloadProgress utils.GenericProgress
	extractProgress  utils.GenericProgress
	downloading      bool
	complete         bool
}

type status struct {
	modProgresses   map[string]modProgress
	installName     string
	overallProgress utils.GenericProgress
	done            bool
}

type apply struct {
	root           components.RootModel
	parent         tea.Model
	error          *components.ErrorComponent
	installChannel chan string
	updateChannel  chan cli.InstallUpdate
	doneChannel    chan bool
	errorChannel   chan error
	cancelChannel  chan bool
	title          string
	status         status
	overall        progress.Model
	sub            progress.Model
	cancelled      bool
}

func NewApply(root components.RootModel, parent tea.Model) tea.Model {
	overall := progress.New(progress.WithSolidFill("118"))
	sub := progress.New(progress.WithSolidFill("202"))

	installChannel := make(chan string)
	updateChannel := make(chan cli.InstallUpdate)
	doneChannel := make(chan bool, 1)
	errorChannel := make(chan error)
	cancelChannel := make(chan bool, 1)

	model := &apply{
		root:    root,
		parent:  parent,
		title:   teaUtils.NonListTitleStyle.MarginTop(1).MarginBottom(1).Render("Applying Changes"),
		overall: overall,
		sub:     sub,
		status: status{
			installName: "",
			done:        false,
		},
		installChannel: installChannel,
		updateChannel:  updateChannel,
		doneChannel:    doneChannel,
		errorChannel:   errorChannel,
		cancelChannel:  cancelChannel,
		cancelled:      false,
	}

	go func() {
		for _, installation := range root.GetGlobal().Installations.Installations {
			installChannel <- installation.Path

			installUpdateChannel := make(chan cli.InstallUpdate)
			go func() {
				for update := range installUpdateChannel {
					updateChannel <- update
				}
			}()

			if err := installation.Install(root.GetGlobal(), installUpdateChannel); err != nil {
				errorChannel <- err
				return
			}

			stop := false
			select {
			case <-cancelChannel:
				stop = true
			default:
			}

			if stop {
				break
			}
		}

		doneChannel <- true
	}()

	return model
}

func (m apply) Init() tea.Cmd {
	return teaUtils.Ticker()
}

func (m apply) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case keys.KeyControlC:
			return m, tea.Quit
		case keys.KeyEscape:
			m.cancelled = true
			m.cancelChannel <- true
			return m, nil
		case keys.KeyEnter:
			if m.status.done {
				if m.parent != nil {
					return m.parent, m.parent.Init()
				}
			}
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.root.SetSize(msg)
	case components.ErrorComponentTimeoutMsg:
		m.error = nil
	case teaUtils.TickMsg:
		select {
		case <-m.doneChannel:
			m.status.done = true
			m.status.installName = ""
			break
		case installName := <-m.installChannel:
			m.status.installName = installName
			m.status.modProgresses = make(map[string]modProgress)
			m.status.overallProgress = utils.GenericProgress{}
			break
		case update := <-m.updateChannel:
			switch update.Type {
			case cli.InstallUpdateTypeOverall:
				m.status.overallProgress = update.Progress
			case cli.InstallUpdateTypeModDownload:
				m.status.modProgresses[update.Item.Mod] = modProgress{
					downloadProgress: update.Progress,
					downloading:      true,
					complete:         false,
				}
			case cli.InstallUpdateTypeModExtract:
				m.status.modProgresses[update.Item.Mod] = modProgress{
					extractProgress: update.Progress,
					downloading:     false,
					complete:        false,
				}
			case cli.InstallUpdateTypeModComplete:
				m.status.modProgresses[update.Item.Mod] = modProgress{
					complete: true,
				}
			}
			break
		case err := <-m.errorChannel:
			wrappedErrMessage := wrap.String(err.Error(), int(float64(m.root.Size().Width)*0.8))
			errorComponent, _ := components.NewErrorComponent(wrappedErrMessage, 0)
			m.error = errorComponent
			break
		default:
			// Skip if nothing there
			break
		}
		return m, teaUtils.Ticker()
	}

	return m, nil
}

func (m apply) View() string {
	strs := make([]string, 0)

	if m.status.installName != "" {
		strs = append(strs, lipgloss.NewStyle().Render(m.status.installName))
		strs = append(strs, lipgloss.NewStyle().MarginBottom(1).Render(m.overall.ViewAs(m.status.overallProgress.Percentage())))
	}

	keys := make([]string, 0)
	for k := range m.status.modProgresses {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, modReference := range keys {
		p := m.status.modProgresses[modReference]
		if p.complete {
			strs = append(strs, lipgloss.NewStyle().Foreground(lipgloss.Color("22")).Render("✓ ")+modReference)
		} else {
			if p.downloading {
				strs = append(strs, lipgloss.NewStyle().Render(modReference+" (Downloading)"))
				strs = append(strs, m.sub.ViewAs(p.downloadProgress.Percentage()))
			} else {
				strs = append(strs, lipgloss.NewStyle().Render(modReference+" (Extracting)"))
				strs = append(strs, m.sub.ViewAs(p.extractProgress.Percentage()))
			}
		}
	}

	if m.status.done {
		if m.cancelled {
			strs = append(strs, teaUtils.LabelStyle.Copy().Foreground(lipgloss.Color("196")).Padding(0).Margin(1).Render("Cancelled! Press Enter to return"))
		} else {
			strs = append(strs, teaUtils.LabelStyle.Copy().Padding(0).Margin(1).Render("Done! Press Enter to return"))
		}
	}

	result := lipgloss.NewStyle().MarginLeft(1).Render(lipgloss.JoinVertical(lipgloss.Left, strs...))

	if m.error != nil {
		return lipgloss.JoinVertical(lipgloss.Left, m.title, m.error.View(), result)
	}

	return lipgloss.JoinVertical(lipgloss.Left, m.title, result)
}
