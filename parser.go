package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
)

// LogParser handles parsing of Ansible log files
type LogParser struct {
	tasks []Task
}

// NewLogParser creates a new LogParser instance
func NewLogParser() *LogParser {
	return &LogParser{
		tasks: make([]Task, 0),
	}
}

// ParseFile parses an Ansible log file and extracts tasks
func (p *LogParser) ParseFile(filename string) ([]Task, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var currentTask *Task
	taskID := 1

	taskRegex := regexp.MustCompile(`^TASK \[(.*?)\] \*+$`)
	startedRegex := regexp.MustCompile(`\[started TASK: (.*?) on (.*?)\]`)
	pathRegex := regexp.MustCompile(`task path: (.*)`)
	// Time format: Tuesday 28 October 2025  02:05:23 +0100
	timeRegex := regexp.MustCompile(`^(\w+) (\d+) (\w+) (\d+)  (\d+):(\d+):(\d+)`)
	
	// Status regexes
	okRegex := regexp.MustCompile(`^ok: \[(.*?)\]`)
	changedRegex := regexp.MustCompile(`^changed: \[(.*?)\]`)
	skippingRegex := regexp.MustCompile(`^skipping: \[(.*?)\]`)
	failedRegex := regexp.MustCompile(`^failed: \[(.*?)\]`)

	// Map month names to numbers for parsing
	monthMap := map[string]string{
		"January": "01", "February": "02", "March": "03", "April": "04",
		"May": "05", "June": "06", "July": "07", "August": "08",
		"September": "09", "October": "10", "November": "11", "December": "12",
	}

	for scanner.Scan() {
		line := scanner.Text()

		// Match task header
		if matches := taskRegex.FindStringSubmatch(line); len(matches) > 1 {
			if currentTask != nil {
				p.tasks = append(p.tasks, *currentTask)
			}
			
			currentTask = &Task{
				ID:          taskID,
				Description: strings.TrimSpace(matches[1]),
				Status:      "unknown", // Default status
			}
			taskID++
			continue
		}

		// If we don't have a current task, skip
		if currentTask == nil {
			continue
		}

		// Extract task path
		if matches := pathRegex.FindStringSubmatch(line); len(matches) > 1 {
			currentTask.Path = matches[1]
			continue
		}

		// Extract start time
		if matches := timeRegex.FindStringSubmatch(line); len(matches) > 7 {
			// Parse the time: Tuesday 28 October 2025  02:05:23
			// weekday := matches[1]  // Not used, commented out
			day := matches[2]
			monthStr := matches[3]
			year := matches[4]
			hour := matches[5]
			minute := matches[6]
			second := matches[7]
			
			// Convert month name to number
			monthNum := monthMap[monthStr]
			if monthNum == "" {
				monthNum = "01" // Default to January
			}
			
			// Format: 2025-10-28 02:05:23
			timeStr := fmt.Sprintf("%s-%s-%s %s:%s:%s", year, monthNum, day, hour, minute, second)
			if t, err := time.Parse("2006-01-02 15:04:05", timeStr); err == nil {
				currentTask.StartTime = t
			}
			continue
		}

		// Extract host from started line
		if matches := startedRegex.FindStringSubmatch(line); len(matches) > 2 {
			// We already set the description from the task header
			continue
		}

		// Check for status updates
		if matches := okRegex.FindStringSubmatch(line); len(matches) > 1 {
			currentTask.Status = "ok"
			currentTask.Host = matches[1]
			continue
		}

		if matches := changedRegex.FindStringSubmatch(line); len(matches) > 1 {
			currentTask.Status = "changed"
			currentTask.Host = matches[1]
			continue
		}

		if matches := skippingRegex.FindStringSubmatch(line); len(matches) > 1 {
			currentTask.Status = "skipping"
			currentTask.Host = matches[1]
			continue
		}

		if matches := failedRegex.FindStringSubmatch(line); len(matches) > 1 {
			currentTask.Status = "failed"
			currentTask.Host = matches[1]
			continue
		}
	}

	// Add the last task if it exists
	if currentTask != nil {
		p.tasks = append(p.tasks, *currentTask)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	return p.tasks, nil
}