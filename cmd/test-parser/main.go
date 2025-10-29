package main

import (
	"fmt"
	"log"
	
	"./.."
)

func main() {
	// Test the parser with the sample log file
	parser := NewLogParser()
	tasks, err := parser.ParseFile("ansible-sample.out")
	if err != nil {
		log.Fatalf("Error parsing file: %v", err)
	}

	fmt.Printf("Found %d tasks:\n", len(tasks))
	for i, task := range tasks {
		if i >= 5 { // Only show first 5 tasks
			fmt.Printf("... and %d more tasks\n", len(tasks)-5)
			break
		}
		fmt.Printf("Task %d: %s - %s (%s)\n", task.ID, task.StartTime.Format("2006-01-02 15:04:05"), task.Description, task.Status)
	}
}