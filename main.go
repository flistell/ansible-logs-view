package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Please provide a log file path as an argument")
	}

	filename := os.Args[1]
	
	parser := NewLogParser()
	tasks, err := parser.ParseFile(filename)
	if err != nil {
		log.Fatalf("Error parsing file: %v", err)
	}

	if len(tasks) == 0 {
		log.Fatal("No tasks found in the log file")
	}

	// Create and run TUI
	m := newModel(tasks)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}
}