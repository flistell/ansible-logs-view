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

	// Open debug log file
	debugFile, err := os.OpenFile("debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("error opening debug log file: %v", err)
	}
	defer debugFile.Close()

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
	
	// Diff regexes
	diffStartRegex := regexp.MustCompile(`^--- before:`)
	
	// Map month names to numbers for parsing
	monthMap := map[string]string{
		"January": "01", "February": "02", "March": "03", "April": "04",
		"May": "05", "June": "06", "July": "07", "August": "08",
		"September": "09", "October": "10", "November": "11", "December": "12",
	}

	// Variables for diff parsing
	inDiffSection := false
	var diffLines []string

	for scanner.Scan() {
		line := scanner.Text()

		// Check if we're entering a new task
		if strings.HasPrefix(line, "TASK [") {
			// If we have a current task, save it
			if currentTask != nil {
				// Add any remaining diff content
				if len(diffLines) > 0 {
					if currentTask.Diff != "" {
						currentTask.Diff += "\n" + strings.Join(diffLines, "\n")
					} else {
						currentTask.Diff = strings.Join(diffLines, "\n")
					}
				}
				// Log the task before appending to tasks
				debugLog := fmt.Sprintf("Task ID: %d\nDescription: %s\nStatus: %s\nHost: %s\nPath: %s\nStartTime: %s\nDiff: %s\nRawText (first 100 chars): %s\n\n",
					currentTask.ID, currentTask.Description, currentTask.Status, currentTask.Host, 
					currentTask.Path, currentTask.StartTime.Format("2006-01-02 15:04:05"), 
					currentTask.Diff, 
					func() string {
						if len(currentTask.RawText) > 100 {
							return currentTask.RawText[:100] + "..."
						}
						return currentTask.RawText
					}())
				debugFile.WriteString(debugLog)
				
				p.tasks = append(p.tasks, *currentTask)
			}
			
			// Reset diff lines for the new task
			diffLines = nil
			
			currentTask = &Task{
				ID:          taskID,
				Description: strings.TrimSpace(taskRegex.FindStringSubmatch(line)[1]),
				Status:      "unknown", // Default status
				RawText:     line + "\n", // Start building raw text with the task header
			}
			taskID++
			continue
		}

		// If we don't have a current task, skip
		if currentTask == nil {
			continue
		}

		// Add the current line to the raw text
		currentTask.RawText += line + "\n"

		// Check if we're entering a diff section
		if diffStartRegex.MatchString(line) {
			inDiffSection = true
			diffLines = []string{line}
			continue
		}
		
		// If we're in a diff section, collect lines until we hit a blank line or task separator
		if inDiffSection {
			// End of diff section when we hit a blank line, task separator, or status line
			if line == "" || strings.HasPrefix(line, "TASK [") || 
			   strings.HasPrefix(line, "ok:") || strings.HasPrefix(line, "changed:") ||
			   strings.HasPrefix(line, "skipping:") || strings.HasPrefix(line, "failed:") {
				// Save the diff to the current task
				if len(diffLines) > 0 {
					if currentTask.Diff != "" {
						currentTask.Diff += "\n" + strings.Join(diffLines, "\n")
					} else {
						currentTask.Diff = strings.Join(diffLines, "\n")
					}
				}
				inDiffSection = false
				diffLines = nil
				
				// If this was a new task, we've already processed it above
				if strings.HasPrefix(line, "TASK [") {
					continue
				}
			} else {
				diffLines = append(diffLines, line)
			}
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
			currentTask.Host = matches[2] // Set the host from the started line
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
		// Add any remaining diff content
		if len(diffLines) > 0 {
			if currentTask.Diff != "" {
				currentTask.Diff += "\n" + strings.Join(diffLines, "\n")
			} else {
				currentTask.Diff = strings.Join(diffLines, "\n")
			}
		}
		// Log the last task
		debugLog := fmt.Sprintf("Task ID: %d\nDescription: %s\nStatus: %s\nHost: %s\nPath: %s\nStartTime: %s\nDiff: %s\nRawText (first 100 chars): %s\n\n",
			currentTask.ID, currentTask.Description, currentTask.Status, currentTask.Host, 
			currentTask.Path, currentTask.StartTime.Format("2006-01-02 15:04:05"), 
			currentTask.Diff, 
			func() string {
				if len(currentTask.RawText) > 1000 {
					return currentTask.RawText[:1000] + "..."
				}
				return currentTask.RawText
			}())
		debugFile.WriteString(debugLog)
		
		p.tasks = append(p.tasks, *currentTask)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	return p.tasks, nil
}