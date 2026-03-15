package main

import (
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/whitehai11/AWaN-TUI/api"
	"github.com/whitehai11/AWaN-TUI/internal/updater"
	"github.com/whitehai11/AWaN-TUI/ui"
)

func main() {
	client := api.NewClient(os.Getenv("AWAN_CORE_URL"))

	updater.StartBackground(updater.Options{
		AppName:        "AWaN TUI",
		Repo:           "whitehai11/AWaN-TUI",
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
