package scenes

import (
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/muesli/reflow/wrap"

	"github.com/satisfactorymodding/ficsit-cli/cli"
	"github.com/satisfactorymodding/ficsit-cli/tea/components"
	"github.com/satisfactorymodding/ficsit-cli/tea/scenes/keys"
	"github.com/satisfactorymodding/ficsit-cli/tea/utils"
)

var _ tea.Model = (*apply)(nil)

type update struct {
	installName    string
	modName        string
	completed      []string
	installTotal   int
	installCurrent int
	modTotal       int
	modCurrent     int
	done           bool
}

type apply struct {
	root          components.RootModel
	parent        tea.Model
	error         *components.ErrorComponent
	updateChannel chan update
	errorChannel  chan error
	cancelChannel chan bool
	title         string
	status        update
	overall       progress.Model
	sub           progress.Model
	cancelled     bool
}

func NewApply(root components.RootModel, parent tea.Model) tea.Model {
	overall := progress.New(progress.WithSolidFill("118"))
	sub := progress.New(progress.WithSolidFill("202"))

	updateChannel := make(chan update)
	errorChannel := make(chan error)
	cancelChannel := make(chan bool, 1)

	model := &apply{
		root:    root,
		parent:  parent,
		title:   utils.NonListTitleStyle.MarginTop(1).MarginBottom(1).Render("Applying Changes"),
		overall: overall,
		sub:     sub,
		status: update{
			completed: []string{},

			installName:    "",
			installTotal:   100,
			installCurrent: 0,

			modName:    "",
			modTotal:   100,
			modCurrent: 0,

			done: false,
		},
		updateChannel: updateChannel,
		errorChannel:  errorChannel,
		cancelChannel: cancelChannel,
		cancelled:     false,
	}

	go func() {
		result := &update{
			completed: make([]string, 0),

			installName:    "",
			installTotal:   100,
			installCurrent: 0,

			modName:    "",
			modTotal:   100,
			modCurrent: 0,

			done: false,
		}
		updateChannel <- *result

		for _, installation := range root.GetGlobal().Installations.Installations {
			result.installName = installation.Path
			updateChannel <- *result

			installChannel := make(chan cli.InstallUpdate)

			go func() {
				for data := range installChannel {
					result.installName = installation.Path
					result.installCurrent = int(data.OverallProgress * 100)

					if data.DownloadProgress < 1 {
						result.modName = "Downloading: " + data.ModName
						result.modCurrent = int(data.DownloadProgress * 100)
					} else {
						result.modName = "Extracting: " + data.ModName
						result.modCurrent = int(data.ExtractProgress * 100)
					}

					updateChannel <- *result
				}
			}()

			if err := installation.Install(root.GetGlobal(), installChannel); err != nil {
				errorChannel <- err
				return
			}

			close(installChannel)

			result.modName = ""
			result.installTotal = 100
			result.completed = append(result.completed, installation.Path)
			updateChannel <- *result

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

		result.done = true
		result.installName = ""
		updateChannel <- *result
	}()

	return model
}

func (m apply) Init() tea.Cmd {
	return utils.Ticker()
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
	case utils.TickMsg:
		select {
		case newStatus := <-m.updateChannel:
			m.status = newStatus
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
		return m, utils.Ticker()
	}

	return m, nil
}

func (m apply) View() string {
	strs := make([]string, 0)
	for _, s := range m.status.completed {
		strs = append(strs, lipgloss.NewStyle().Foreground(lipgloss.Color("22")).Render("✓ ")+s)
	}

	if m.status.installName != "" {
		marginTop := 0
		if len(m.status.completed) > 0 {
			marginTop = 1
		}

		strs = append(strs, lipgloss.NewStyle().MarginTop(marginTop).Render(m.status.installName))
		strs = append(strs, m.overall.ViewAs(float64(m.status.installCurrent)/float64(m.status.installTotal)))
	}

	if m.status.modName != "" {
		strs = append(strs, lipgloss.NewStyle().MarginTop(1).Render(m.status.modName))
		strs = append(strs, m.sub.ViewAs(float64(m.status.modCurrent)/float64(m.status.modTotal)))
	}

	if m.status.done {
		if m.cancelled {
			strs = append(strs, utils.LabelStyle.Copy().Foreground(lipgloss.Color("196")).Padding(0).Margin(1).Render("Cancelled! Press Enter to return"))
		} else {
			strs = append(strs, utils.LabelStyle.Copy().Padding(0).Margin(1).Render("Done! Press Enter to return"))
		}
	}

	result := lipgloss.NewStyle().MarginLeft(1).Render(lipgloss.JoinVertical(lipgloss.Left, strs...))

	if m.error != nil {
		return lipgloss.JoinVertical(lipgloss.Left, m.title, m.error.View(), result)
	}

	return lipgloss.JoinVertical(lipgloss.Left, m.title, result)
}
