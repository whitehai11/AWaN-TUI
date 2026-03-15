package main

import (
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/awan/awan-tui/api"
	"github.com/awan/awan-tui/internal/updater"
	"github.com/awan/awan-tui/ui"
)

func main() {
	client := api.NewClient(os.Getenv("AWAN_CORE_URL"))

	updater.StartBackground(updater.Options{
		AppName:        "AWaN TUI",
		Repo:           "awan/tui",
		Version:        Version,
		BinaryBaseName: "awan-tui",
		Args:           os.Args[1:],
		Logger: func(message string) {
			log.Println("[AWAN]", message)
		},
	})

	program := tea.NewProgram(
		ui.NewModel(client),
		tea.WithAltScreen(),
	)

	if _, err := program.Run(); err != nil {
		log.Fatal(err)
	}
}
