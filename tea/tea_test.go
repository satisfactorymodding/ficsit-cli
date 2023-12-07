package tea

import (
	"io"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/MarvinJWendt/testza"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/pkg/errors"

	"github.com/satisfactorymodding/ficsit-cli/cfg"
	"github.com/satisfactorymodding/ficsit-cli/cli"
	"github.com/satisfactorymodding/ficsit-cli/tea/scenes"
)

func init() {
	cfg.SetDefaults()
}

func TestTea(t *testing.T) {
	if runtime.GOOS == "windows" {
		// Windows just sucks
		return
	}

	serverLocation := os.Getenv("SF_DEDICATED_SERVER")
	if serverLocation == "" {
		return
	}

	ctx, err := cli.InitCLI(false)
	testza.AssertNoError(t, err)

	ctx.Provider = cli.MockProvider{}

	err = ctx.Wipe()
	testza.AssertNoError(t, err)

	err = ctx.ReInit()
	testza.AssertNoError(t, err)

	root := newModel(ctx)
	m := scenes.NewMainMenu(root)

	tm := teatest.NewTestModel(
		t, m,
		teatest.WithInitialTermSize(70, 35),
	)

	t.Cleanup(func() {
		if err := tm.Quit(); err != nil {
			t.Fatal(err)
		}
	})

	time.Sleep(time.Second)

	// Go to Installations
	press(tm, tea.KeyEnter)

	// Create new installation
	write(tm, "n")

	// Enter installation path
	write(tm, serverLocation)

	// Accept path
	press(tm, tea.KeyEnter)

	// Go back to main menu
	write(tm, "q")

	// Go to all mods
	press(tm, tea.KeyDown)
	press(tm, tea.KeyDown)
	press(tm, tea.KeyDown)
	press(tm, tea.KeyEnter)

	// Filter for mod
	write(tm, "/")
	write(tm, "Refined Power")
	press(tm, tea.KeyEnter)

	// Select mod
	press(tm, tea.KeyEnter)

	// Install mod
	press(tm, tea.KeyEnter)

	// Go back to main menu
	write(tm, "q")

	// Apply changes
	press(tm, tea.KeyDown)
	press(tm, tea.KeyDown)
	press(tm, tea.KeyDown)

	eat(tm)
	press(tm, tea.KeyEnter)

	i := 0
	buffer := ""
	for {
		s := read(tm)
		buffer += "\n-------------------------\n" + s

		if strings.Contains(s, "Done! Press Enter to return") {
			break
		}

		if strings.Contains(s, "Cancelled! Press Enter to return") {
			testza.AssertNoError(t, errors.New("installation cancelled"))
			println(buffer)
			break
		}

		i++
		if i >= 60 {
			testza.AssertNoError(t, errors.New("failed installing"))
			println(buffer)
			return
		}

		time.Sleep(time.Second)
	}

	eat(tm)

	// Go back to main menu
	press(tm, tea.KeyEnter)

	// Exit program
	press(tm, tea.KeyDown)
	press(tm, tea.KeyDown)
	press(tm, tea.KeyEnter)
}

// dump the current tea buffer to stderr
func dump(tm *teatest.TestModel) { // nolint
	_, _ = io.Copy(os.Stderr, tm.Output())
}

// eat the current tea buffer
func eat(tm *teatest.TestModel) {
	_, _ = io.ReadAll(tm.Output())
}

// read reads the current tea buffer
func read(tm *teatest.TestModel) string {
	out, _ := io.ReadAll(tm.Output())
	return string(out)
}

func press(tm *teatest.TestModel, key tea.KeyType) {
	println("Pressing", key.String())
	tm.Send(tea.KeyMsg{Type: key})
	time.Sleep(time.Millisecond * 250)
}

func write(tm *teatest.TestModel, txt string) {
	println("Writing", txt)
	tm.Type(txt)
	time.Sleep(time.Millisecond * 250)
}
