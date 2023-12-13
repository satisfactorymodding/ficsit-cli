package scenes

import (
	"sort"
	"sync"

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
	root          components.RootModel
	parent        tea.Model
	error         *components.ErrorComponent
	updateChannel chan applyUpdate
	doneChannel   chan bool
	errorChannel  chan error
	cancelChannel chan bool
	title         string
	status        map[string]status
	overall       progress.Model
	sub           progress.Model
	cancelled     bool
	done          bool
}

type applyUpdate struct {
	Installation *cli.Installation
	Update       cli.InstallUpdate
	Done         bool
}

func NewApply(root components.RootModel, parent tea.Model) tea.Model {
	overall := progress.New(progress.WithSolidFill("118"))
	sub := progress.New(progress.WithSolidFill("202"))

	updateChannel := make(chan applyUpdate)
	doneChannel := make(chan bool, 1)
	errorChannel := make(chan error)
	cancelChannel := make(chan bool, 1)

	model := &apply{
		root:          root,
		parent:        parent,
		title:         teaUtils.NonListTitleStyle.MarginTop(1).MarginBottom(1).Render("Applying Changes"),
		overall:       overall,
		sub:           sub,
		status:        make(map[string]status),
		updateChannel: updateChannel,
		doneChannel:   doneChannel,
		errorChannel:  errorChannel,
		cancelChannel: cancelChannel,
	}

	var wg sync.WaitGroup

	for _, installation := range root.GetGlobal().Installations.Installations {
		wg.Add(1)

		model.status[installation.Path] = status{
			modProgresses:   make(map[string]modProgress),
			installName:     installation.Path,
			overallProgress: utils.GenericProgress{},
		}

		go func(installation *cli.Installation) {
			defer wg.Done()

			installUpdateChannel := make(chan cli.InstallUpdate)
			go func() {
				for update := range installUpdateChannel {
					updateChannel <- applyUpdate{
						Installation: installation,
						Update:       update,
					}
				}
			}()

			if err := installation.Install(root.GetGlobal(), installUpdateChannel); err != nil {
				errorChannel <- err
				return
			}

			updateChannel <- applyUpdate{
				Installation: installation,
				Done:         true,
			}
		}(installation)
	}

	go func() {
		wg.Wait()
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
		case keys.KeyQ:
			fallthrough
		case keys.KeyEscape:
			if m.done {
				if m.parent != nil {
					return m.parent, m.parent.Init()
				}
				return m, tea.Quit
			}

			m.cancelled = true

			if m.error != nil {
				if m.parent != nil {
					return m.parent, m.parent.Init()
				}
				return m, tea.Quit
			}

			m.cancelChannel <- true
			return m, nil
		case keys.KeyEnter:
			if m.done || m.error != nil {
				if m.parent != nil {
					return m.parent, m.parent.Init()
				}
				return m, tea.Quit
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
			m.done = true
			break
		case update := <-m.updateChannel:
			s := m.status[update.Installation.Path]

			if update.Done {
				s.done = true
			} else {
				switch update.Update.Type {
				case cli.InstallUpdateTypeOverall:
					s.overallProgress = update.Update.Progress
				case cli.InstallUpdateTypeModDownload:
					s.modProgresses[update.Update.Item.Mod] = modProgress{
						downloadProgress: update.Update.Progress,
						downloading:      true,
						complete:         false,
					}
				case cli.InstallUpdateTypeModExtract:
					s.modProgresses[update.Update.Item.Mod] = modProgress{
						extractProgress: update.Update.Progress,
						downloading:     false,
						complete:        false,
					}
				case cli.InstallUpdateTypeModComplete:
					s.modProgresses[update.Update.Item.Mod] = modProgress{
						complete: true,
					}
				}
			}

			m.status[update.Installation.Path] = s
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

	installationList := make([]string, len(m.status))
	i := 0
	for key := range m.status {
		installationList[i] = key
		i++
	}

	sort.Strings(installationList)

	totalHeight := 3 + 3                     // Header + Footer
	totalHeight += len(installationList) * 2 // Bottom Margin + Overall progress per-install

	bottomMargins := 1
	if m.root.Size().Height < totalHeight {
		bottomMargins = 0
	}

	totalHeight += len(installationList) // Top margin

	topMargins := 1
	if m.root.Size().Height < totalHeight {
		topMargins = 0
	}

	for _, installPath := range installationList {
		totalHeight += len(m.status[installPath].modProgresses)
	}

	for _, installPath := range installationList {
		s := m.status[installPath]

		strs = append(strs, lipgloss.NewStyle().Margin(topMargins, 0, bottomMargins, 1).Render(lipgloss.JoinHorizontal(
			lipgloss.Left,
			m.overall.ViewAs(s.overallProgress.Percentage()),
			" - ",
			lipgloss.NewStyle().Render(installPath),
		)))

		modReferences := make([]string, 0)
		for k := range s.modProgresses {
			modReferences = append(modReferences, k)
		}
		sort.Strings(modReferences)

		if m.root.Size().Height > totalHeight {
			for _, modReference := range modReferences {
				p := s.modProgresses[modReference]
				if p.complete || s.done {
					strs = append(strs, lipgloss.NewStyle().Foreground(lipgloss.Color("22")).Render("✓ ")+modReference)
				} else {
					if p.downloading {
						strs = append(strs, lipgloss.JoinHorizontal(
							lipgloss.Left,
							m.sub.ViewAs(p.downloadProgress.Percentage()),
							" - ",
							lipgloss.NewStyle().Render(modReference+" (Downloading)"),
						))
					} else {
						strs = append(strs, lipgloss.JoinHorizontal(
							lipgloss.Left,
							m.sub.ViewAs(p.extractProgress.Percentage()),
							" - ",
							lipgloss.NewStyle().Render(modReference+" (Extracting)"),
						))
					}
				}
			}
		}
	}

	if m.done {
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
