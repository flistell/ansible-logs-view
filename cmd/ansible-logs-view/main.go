package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"ansible-logs-view/internal/app"

	tea "github.com/charmbracelet/bubbletea"
)


func main() {
	debug := flag.Bool("debug", false, "Enable debug logging to debug.log")
	flag.Parse()

	if len(flag.Args()) < 1 {
		log.Fatal("Please provide a log file path as an argument")
	}

	filename := flag.Args()[0]
	
	parser := app.NewLogParser(*debug)
	tasks, err := parser.ParseFile(filename)
	if err != nil {
		log.Fatalf("Error parsing file: %v", err)
	}

	if len(tasks) == 0 {
		log.Fatal("No tasks found in the log file")
	}

	// Create and run TUI
	m := app.NewModel(tasks, *debug)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}
}