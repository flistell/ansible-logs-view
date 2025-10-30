# Ansible Logs TUI

A terminal user interface (TUI) for viewing and analyzing Ansible log files, built with Go and Bubbletea.

![Demo](demo.gif)

## Features

- Parse Ansible log files to extract tasks
- Display tasks in a scrollable list with their status (ok, changed, skipping, failed)
- Navigate through tasks using keyboard controls
- Expand/collapse tasks to view detailed information without changing panels
- View full raw task text in a separate details panel when a task is expanded
- Details panel height fixed to 1/3 of the screen
- Details panel width fixed to full screen width
- Raw text content in details panel wraps if lines are too long
- Details panel is scrollable with PgUp/PgDn keys
- Color-coded status indicators for quick visual identification
- Filter tasks by description, status, date, host, path, or diff content
- Debug logging of task structure to debug.log file

## Installation

1. Clone the repository:
   ```
   git clone <repository-url>
   cd ansible-logs-tui
   ```

2. Build the application:
   ```
   go build -o ansible-log-view
   ```

## Usage

Run the application with an Ansible log file as an argument:
```
./ansible-log-view /path/to/ansible-log-file.log
```

### Keyboard Controls

- `↑` / `↓` : Navigate through tasks
- `Enter` / `Space` : Expand/collapse selected task and show full raw task text in separate panel
- `PgUp` / `PgDn` : Scroll details panel when visible
- `g` : Go to the top of the task list
- `G` : Go to the bottom of the task list
- `/` : Toggle filter input
- `q` / `Ctrl+C` : Quit the application

### Filtering Tasks

1. Press `/` to open the filter input
2. Type your search term (any part of task description, status, date, host, path, or diff content)
3. Press `Enter` to apply the filter
4. Press `Esc` to cancel filtering and restore all tasks

### Viewing Task Details and Diffs

1. Navigate to a task using arrow keys
2. Press `Enter` or `Space` to expand the selected task
3. View full raw task text in the separate details panel at the bottom (fixed to 1/3 of screen height)
4. Use `PgUp`/`PgDn` to scroll through long content in the details panel
5. Press `Enter` or `Space` again to collapse the task and hide details panel

### Debug Logging

The application now creates a `debug.log` file that contains detailed information about each parsed task, including:
- Task ID
- Description
- Status
- Host
- Path
- Start time
- Diff information
- First 100 characters of RawText

## Development

### Dependencies

- Go 1.16+
- Bubbletea v1.3.10
- Bubbles v0.21.0

### Project Structure

- `main.go` : Entry point of the application
- `parser.go` : Parses Ansible log files and extracts tasks
- `task.go` : Defines the Task struct
- `tui.go` : Implements the terminal user interface

### Building

To build the application:
```
go build -o ansible-log-view
```

### Running Tests

To run tests:
```
go test ./...
```

## Technical Details

### Log Format Understanding

Through careful analysis of the Ansible log file, the following patterns were identified:
- Tasks begin with `TASK [description] *****` headers
- Each task contains metadata including timestamps and paths
- Task execution status is indicated by lines like `ok:`, `changed:`, `skipping:`, or `failed:`
- Timestamps follow the format: `DayOfWeek Day Month Year HH:MM:SS`
- Diff information appears in sections starting with `--- before:`

### UI/UX Design Decisions

- **Color Coding**: Different status types are visually distinguished with appropriate colors:
  - Green for successful tasks (`ok`)
  - Orange for tasks that made changes (`changed`)
  - Gray for skipped tasks (`skipping`)
  - Red for failed tasks (`failed`)
- **Responsive Layout**: The interface adapts to terminal window size changes
- **Clear Navigation**: Intuitive keyboard controls with visual feedback
- **Dual-Panel Display**: Task list and raw task text shown simultaneously
- **Viewport Scrolling**: Efficient handling of large numbers of tasks
- **Search/Filter**: Quick access to specific tasks by keyword

## Benefits

### For DevOps Engineers
- **Quick Analysis**: Rapidly identify which tasks executed successfully, changed systems, or failed
- **Change Tracking**: Easily see exactly what files were modified by each task
- **Troubleshooting**: Quickly pinpoint problematic tasks and understand their impact
- **Task Filtering**: Find specific tasks by description, status, date, or other criteria

### For System Administrators
- **Deployment Verification**: Confirm that deployments executed as expected
- **Audit Trail**: Maintain a clear record of system changes
- **Issue Resolution**: Speed up debugging by focusing on specific task changes

### For Developers
- **Learning Tool**: Understand how Ansible tasks affect system state
- **Code Review**: Examine the actual changes made by deployment scripts
- **Documentation**: Use the tool to document deployment behaviors

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.