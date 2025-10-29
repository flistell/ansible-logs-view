# Ansible Logs TUI

A terminal user interface (TUI) for viewing Ansible log files, built with Go and Bubbletea.

## Features

- Parse Ansible log files to extract tasks
- Display tasks in a tree view with their status (ok, changed, skipping, failed)
- Navigate through tasks using keyboard controls
- Expand/collapse tasks to view detailed information without changing panels

## Installation

1. Clone the repository:
   ```
   git clone <repository-url>
   cd ansible-logs-tui
   ```

2. Build the application:
   ```
   go build
   ```

## Usage

Run the application with an Ansible log file as an argument:
```
./ansible-logs-tui /path/to/ansible-log-file.log
```

### Keyboard Controls

- `↑` / `↓` : Navigate through tasks
- `Enter` / `Space` : Expand/collapse selected task
- `q` / `Esc` / `Ctrl+C` : Quit the application

## Development

### Dependencies

- Go 1.16+
- Bubbletea
- Bubbles

### Project Structure

- `main.go` : Entry point of the application
- `parser.go` : Parses Ansible log files and extracts tasks
- `task.go` : Defines the Task struct
- `tui.go` : Implements the terminal user interface

### Building

To build the application:
```
go build
```

### Running Tests

To run tests:
```
go test ./...
```