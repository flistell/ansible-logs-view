package main

import (
	"bufio"
	"os"
	"strings"
)

type TestItem struct {
	Title   string
	Details string
}

// readTestItems reads items from the test data file
func readTestItems(filepath string) ([]TestItem, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var items []TestItem
	var currentItem *TestItem
	scanner := bufio.NewScanner(file)
	
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "- title:") {
			if currentItem != nil {
				items = append(items, *currentItem)
			}
			title := strings.TrimPrefix(line, "- title:")
			currentItem = &TestItem{
				Title: strings.TrimSpace(title),
			}
		} else if strings.HasPrefix(line, "details:") && currentItem != nil {
			details := strings.TrimPrefix(line, "details:")
			currentItem.Details = strings.TrimSpace(details)
		}
	}

	// Add the last item
	if currentItem != nil {
		items = append(items, *currentItem)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return items, nil
}