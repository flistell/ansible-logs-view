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
   cd ansible-logs-view
   ```

2. Build the application:
   ```
   make build
   ```

3. Build the application linked with older glic-2.28. This would require `podman` installed and working on the system.

```
   make build-glibc-2.28
```

## Usage

Run the application with an Ansible log file as an argument:
```
./ansible-logs-view /path/to/ansible-log-file.log
```

Or run with debug mode enabled:
```
./ansible-logs-view --debug /path/to/ansible-log-file.log
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
- First 1000 characters of RawText

To enable debug logging, run the application with the `--debug` flag:
```
./ansible-logs-view --debug /path/to/ansible-log-file.log
```

## Development

### Dependencies

- Go 1.25.3+
- Bubbletea v1.3.10
- Bubbles v0.21.0

### Project Structure

```
ansible-logs-view/
├── go.mod
├── go.sum
├── cmd/
│   ├── ansible-logs-view/
│   │   ├── main.go              # Application entry point
│   │   └── ...
│   └── tui-poc/
│       ├── itemreader.go
│       └── main.go              # Proof of concept TUI (not the main app)
├── internal/
│   └── app/
│       ├── logger.go            # Logging setup
│       ├── parser_test.go       # Parser tests
│       ├── parser.go            # Log file parsing logic
│       ├── task.go              # Task data structure
│       └── tui.go               # Terminal user interface implementation
└── testdata/
    ├── sample.log
    └── testitems.txt
```

### Building

To build the application:
```
go build -o ansible-logs-view ./cmd/ansible-logs-view
```

You can build the tool for an older RHEL/OL/Rockylinux with glibc-2.28 using `./build-glibc-2.28.sh`. To run this script you need `podman`. It will download a rockylinux:8 image, build the tool there, spin a container pull the compiled binary to your host machine.

### Running Tests

To run tests:
```
go test ./...
```

### Key Components

#### 1. Log Parser (`internal/app/parser.go`)
- Parses Ansible log files to extract individual tasks
- Extracts task metadata including:
  - Task ID
  - Description
  - Start time
  - Status (ok, changed, skipping, failed)
  - Host
  - Path
  - Diff information
  - Raw task text from the log file
- Debug logging: Creates debug.log file with detailed information about each parsed task

#### 2. Data Model (`internal/app/task.go`)
- Defines the `Task` struct to represent parsed tasks
- Contains all relevant task information for display including diff data and raw task text

#### 3. Logger (`internal/app/logger.go`)
- Centralized logging implementation for debug output
- Manages debug log file creation and writing
- Thread-safe logger initialization

#### 4. Terminal UI (`internal/app/tui.go`)
- Implements a scrollable list view with viewport-based scrolling
- Provides dual-panel display:
  - **Task List Panel**: Shows all tasks in a scrollable list with color-coded status indicators
  - **Details Panel**: Displays full raw task text for expanded tasks
- Details panel properties:
  - Fixed height to 1/3 of the screen
  - Fixed width to full screen width
  - Content wraps if lines are too long
  - Scrollable with PgUp/PgDn keys
- Provides intuitive keyboard navigation:
  - Arrow keys for navigation
  - Enter/Space to expand/collapse tasks and show full raw task text in separate panel
  - PgUp/PgDn to scroll details panel when visible
  - g/G for top/bottom navigation
  - `/` for filtering tasks
  - Q/Ctrl+C to quit
- Filtering capability to search tasks by description, status, date, host, path, or diff content
- Color-coded status indicators for quick visual identification

#### 5. Main Application (`cmd/ansible-logs-view/main.go`)
- Handles command-line argument parsing
- Initializes the parser and TUI components
- Manages the application lifecycle
- Supports a `--debug` flag to enable debug logging

#### 6. Parser Tests (`internal/app/parser_test.go`)
- Contains integration-style tests for the parser
- Verifies that the parser correctly extracts task information
- Tests against sample log files to ensure correctness

## Technical Details

### Log Format Understanding

Through careful analysis of the Ansible log file, the following patterns were identified:
- Tasks begin with `TASK [description] *****` headers
- Each task contains metadata including timestamps and paths
- Task execution status is indicated by lines like `ok:`, `changed:`, `skipping:`, or `failed:`
- Timestamps follow the format: `DayOfWeek Day Month Year HH:MM:SS`
- Diff information appears in sections starting with `--- before:`
- Tasks with changes include detailed diff output showing before/after comparisons

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
