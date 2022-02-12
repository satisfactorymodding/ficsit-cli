package scenes

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/satisfactorymodding/ficsit-cli/ficsit"
	"github.com/satisfactorymodding/ficsit-cli/tea/components"
	"github.com/satisfactorymodding/ficsit-cli/tea/utils"
)

var _ tea.Model = (*apply)(nil)

type apply struct {
	root     components.RootModel
	viewport viewport.Model
	spinner  spinner.Model
	parent   tea.Model
	ready    bool
	log      chan string
}

func NewApply(root components.RootModel, parent tea.Model) tea.Model {
	model := apply{
		root: root,
		viewport: viewport.Model{},
		spinner:  spinner.NewModel(),
		parent:   parent,
		ready:    false,
		log:      make(chan string),
	}

	model.spinner.Spinner = spinner.MiniDot

	go func() {
		profile := root.GetGlobal().Profiles.GetProfile(root.GetGlobal().Profiles.SelectedProfile)
		installation := root.GetGlobal().Installations.SelectedInstallation
		modLog := fmt.Sprintf("Installing mods from profile %s\n", profile.Name)

		for _, mod := range profile.Mods {
			modLog += fmt.Sprintf("%s (%s): %s\n", mod.Name, mod.ID, mod.Version)
			model.log <- modLog

			versions, err := ficsit.ModVersions(context.TODO(), root.GetAPIClient(), mod.ID, ficsit.VersionFilter{
				Limit:    100,
				Offset:   0,
				Search:   mod.Version,
				Order:    ficsit.OrderDesc,
				Order_by: ficsit.VersionFieldsCreatedAt,
			})

			if err != nil {
				panic(err) // TODO
			}

			modCount := len(versions.GetMod.Versions)
			if modCount == 0 {
				modLog += "No versions found"
				continue
			}

			version := versions.GetMod.Versions[0]
			url := fmt.Sprintf("https://api.ficsit.app/v1/mod/%s/versions/%s/download", mod.ID, version.Id)
			modPath := path.Join(installation, mod.Reference)
			modZipPath := modPath + ".smod"

			modLog += fmt.Sprintf("Downloading %s %s to %s\n", mod.Name, version.Version, modZipPath)
			model.log <- modLog

			// Get the data
			resp, err := http.Get(url)
			if err != nil {
				modLog += fmt.Sprintf("Could not download: %s\n", err)
				model.log <- modLog
				continue
			}
			defer resp.Body.Close()

			// Create the file
			out, err := os.Create(modZipPath)
			if err != nil {
				modLog += fmt.Sprintf("Could not create output file %s:]n %s\n", modZipPath, err)
				model.log <- modLog
				continue
			}
			defer out.Close()

			// Write the body to file
			_, err = io.Copy(out, resp.Body)
			if err != nil {
				modLog += fmt.Sprintf("Failed to write output file %s:\n %s\n", modZipPath, err)
				model.log <- modLog
				continue
			}

			modLog += "Download complete\n"
			model.log <- modLog

			err = os.MkdirAll(modPath, 0775)
			if err != nil {
				modLog += fmt.Sprintf("Failed to create mod folder %s:\n %s\n", modPath, err)
				model.log <- modLog
				continue
			}

			// Open a zip archive for reading.
			r, err := zip.OpenReader(modZipPath)
			if err != nil {
				modLog += fmt.Sprintf("Failed to open %s:\n %s\n", modZipPath, err)
				model.log <- modLog
				continue
			}
			defer r.Close()

			for _, f := range r.File {
				modLog += fmt.Sprintf("Unzipping %s\n", f.Name)
				pluginFile, err := f.Open()
				if err != nil {
					modLog += fmt.Sprintf("Unable to open %s in %s:\n %s\n", f.Name, modZipPath, err)
				}

				destPath := path.Join(modPath, f.Name)
				destDir := path.Dir(destPath)
				err = os.MkdirAll(destDir, 0775)
				if err != nil {
					modLog += fmt.Sprintf("Failed to create mod subfolder %s:\n %s\n", modPath, err)
					model.log <- modLog
					continue
				}
	
				out, err := os.Create(destPath)
				if err != nil {
					modLog += fmt.Sprintf("Could not create output file %s:\n %s\n\n", destPath, err)
					model.log <- modLog
					continue
				}
	
				io.Copy(out, pluginFile)
			}
			modLog += "Done"
			model.log <- modLog

		}
	}()

	return model
}

func (m apply) Init() tea.Cmd {
	return tea.Batch(utils.Ticker(), spinner.Tick)
}

func (m apply) CalculateSizes(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	if m.viewport.Width == 0 {
		return m, nil
	}

	bottomPadding := 2

	top, right, bottom, left := lipgloss.NewStyle().Margin(m.root.Height(), 3, bottomPadding).GetMargin()
	m.viewport.Width = msg.Width - left - right
	m.viewport.Height = msg.Height - top - bottom
	m.root.SetSize(msg)

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m apply) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		case log := <-m.log:
			bottomPadding := 2
			logData := lipgloss.NewStyle().Padding(0, 2).Render(log)

			top, right, bottom, left := lipgloss.NewStyle().Margin(m.root.Height(), 3, bottomPadding).GetMargin()
			m.viewport = viewport.Model{Width: m.root.Size().Width - left - right, Height: m.root.Size().Height - top - bottom}
			m.viewport.SetContent(lipgloss.JoinVertical(lipgloss.Left, logData))

			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			return m, tea.Batch(cmd, utils.Ticker())
		default:
			return m, utils.Ticker()
		}
	}

	return m, nil
}

func (m apply) View() string {
	spinnerView := lipgloss.NewStyle().Padding(0, 2, 1).Render(m.spinner.View() + " Applying changes...")
	return lipgloss.JoinVertical(lipgloss.Left, m.root.View(), spinnerView, m.viewport.View())
}
