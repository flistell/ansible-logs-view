# Ansible Logs TUI - Project Summary

## Project Overview

This project implements a Terminal User Interface (TUI) application for viewing and analyzing Ansible log files. Built with Go and the Bubbletea framework, it provides an intuitive way to navigate through Ansible tasks and examine their details.

## Task List & Status

### Completed Tasks ✅

1. **Project Initialization**
   - Initialize Go project with go.mod
   - Install Bubbletea library dependency

2. **Log Analysis & Parsing**
   - Analyze ansible-sample.out file structure to understand log format
   - Create log parsing logic to extract tasks, start times, descriptions, and status
   - Enhanced parser to extract diff information from logs
   - Enhanced parser to store raw task text from log files
   - Enhanced parser to add debug logging of currentTask structure to debug.log file

3. **TUI Implementation**
   - Create TUI model with list of tasks
   - Implement scrollable list view showing task description, start time and status
   - Implement detail view showing full task information when selected
   - Add keyboard navigation (up/down to select, enter to expand)
   - Enhanced with diff visualization in a separate panel
   - Enhanced with filtering capability by any part of task description
   - Enhanced to show full raw task text in details panel when 'return' is pressed
   - Fixed details panel to 1/3 height of screen
   - Fixed details panel to full width of screen
   - Added line wrapping for long text in details panel
   - Added PgUp/PgDn scrolling for details panel
   - Create main function to read log file from command argument and run TUI
   - Test the application with ansible-sample.out

## Project Structure

```
ansible-logs-view/
├── .gemini-env
├── .gitignore
├── ansible-sample.out
├── Dockerfile-glibc-2.28
├── GEMINI-REVIEW.md
├── go.mod
├── go.sum
├── QWEN.md
├── README.md
├── .aider.tags.cache.v4/
├── .git/
├── .github/
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

## Key Components

### 1. Log Parser (`internal/app/parser.go`)
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

### 2. Data Model (`internal/app/task.go`)
- Defines the `Task` struct to represent parsed tasks
- Contains all relevant task information for display including diff data and raw task text

### 3. Logger (`internal/app/logger.go`)
- Centralized logging implementation for debug output
- Manages debug log file creation and writing
- Thread-safe logger initialization

### 4. Terminal UI (`internal/app/tui.go`)
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

### 5. Main Application (`cmd/ansible-logs-view/main.go`)
- Handles command-line argument parsing
- Initializes the parser and TUI components
- Manages the application lifecycle
- Supports a `--debug` flag to enable debug logging

### 6. Parser Tests (`internal/app/parser_test.go`)
- Contains integration-style tests for the parser
- Verifies that the parser correctly extracts task information
- Tests against sample log files to ensure correctness

## Technical Insights

### Log Format Understanding
Through careful analysis of the Ansible log file, I identified the following patterns:
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

### Details Panel Enhancement
- **Fixed Dimensions**: Details panel height fixed to 1/3 of screen, width fixed to full screen
- **Content Wrapping**: Text in details panel wraps when lines are too long
- **Page Navigation**: Details panel is scrollable with PgUp/PgDn keys when expanded
- **Viewport Management**: Separate viewport specifically for scrolling the raw text content

### Debug Logging Enhancement
- **Logging Implementation**: Added debug logging to parser.go that creates debug.log file
- **Structured Output**: Logs complete task structure for each parsed task including ID, Description, Status, Host, Path, StartTime, Diff, and RawText
- **Truncation**: RawText is truncated to first 100 characters in logs to prevent log file from becoming too large
- **File Management**: Creates or appends to debug.log file during parsing process

### Raw Task Text Enhancement
- **Parser Enhancement**: Modified the log parser to capture the full raw task text from the log file
- **Task Struct Extension**: Added RawText field to store the complete task entry from the log
- **Details Panel Implementation**: Show the full raw task text when 'return' is pressed on a node
- **Content Display**: Display the complete task content exactly as it appears in the log file

### Diff Visualization Enhancement
- **Parser Enhancement**: Modified the log parser to extract diff information from Ansible logs
- **Diff Detection**: Identifies sections starting with `--- before:` and captures the entire diff block
- **Task Association**: Associates diff information with the corresponding task
- **Dual-Panel Display**: Implements a task list panel and a separate diff panel
- **Space Management**: Dynamically adjusts panel sizes based on available terminal space
- **Visual Design**: Maintains existing styling with color-coded status indicators and adds distinctive styling for the diff panel

### Filtering Enhancement
- **Text Input Integration**: Added text input field for filter terms
- **Comprehensive Search**: Search across task description, status, host, path, timestamp, diff content, and raw text
- **Dynamic Filtering**: Updates results in real-time as user types
- **Keyboard Shortcuts**: `/` to open filter, Enter to apply, Esc to cancel

## Lessons Learned

### Technical Skills
1. **Go Programming**: Deepened my understanding of Go modules, interfaces, and structuring larger applications
2. **Bubbletea Framework**: Gained proficiency with the Bubbletea TUI framework and its component ecosystem (bubbles)
3. **Text Processing**: Improved skills in parsing structured text files and extracting meaningful information
4. **UI/UX Design**: Learned to think about terminal user experience and visual hierarchy in text-based interfaces
5. **Viewport Management**: Mastered Bubbletea's viewport component for efficient scrolling through large datasets

### Development Practices
1. **Incremental Development**: Breaking the project into small, manageable tasks proved highly effective
2. **Iterative Testing**: Regularly testing with the sample log file helped identify parsing edge cases early
3. **Code Organization**: Structuring the application into logical components (parser, model, view) improved maintainability
4. **Error Handling**: Proper error handling throughout the application ensures robust behavior with malformed inputs
5. **Debugging**: Adding debug logging provides valuable insight into parsing behavior
6. **Backward Compatibility**: Maintaining compatibility while adding new features

### Challenges Overcome
1. **Time Parsing Complexity**: Converting natural language dates (e.g., "Tuesday 28 October 2025") to Go time objects required careful string manipulation
2. **Terminal Compatibility**: Ensuring the TUI works well across different terminal sizes and environments
3. **Memory Efficiency**: Handling potentially large log files without excessive memory consumption
4. **Viewport Scrolling**: Implementing smooth scrolling through large task lists using Bubbletea's viewport component
5. **Diff Extraction**: Identifying and extracting diff information from complex log formats
6. **Filtering Implementation**: Adding comprehensive search functionality that filters across multiple fields
7. **Raw Task Text Integration**: Creating a system to capture and display the complete original task text from logs
8. **Debug Logging**: Implementing proper logging to debug task structure without impacting performance
9. **Details Panel Configuration**: Setting fixed dimensions and scrollable behavior for the details panel

## Future Enhancements

While the current implementation is functional and feature-complete, several enhancements could be considered:
1. **Advanced Filtering**: Add ability to filter tasks by specific criteria
2. **Search Functionality**: Implement text search within tasks and diffs
3. **Export Options**: Allow exporting task lists or diffs to various formats
4. **Performance Optimization**: For extremely large log files, implement pagination or streaming
5. **Enhanced Detail View**: Include additional metadata and better formatting of task details
6. **Multi-File Support**: Compare diffs across multiple log files
7. **Configuration**: Add configuration options for colors, display formats, etc.

## Conclusion

This project successfully demonstrates how to build a practical TUI application for technical log analysis. The combination of Go's efficiency and Bubbletea's TUI capabilities creates a powerful tool for DevOps professionals who need to quickly analyze Ansible deployment logs.

The application transforms raw, verbose log files into an interactive, navigable interface that makes troubleshooting and auditing significantly more efficient. With the addition of raw task text display, filtering capabilities, diff visualization, and debug logging, users can now see exactly what changes were made by each task, quickly find specific tasks of interest, view the complete original task content, and troubleshoot parsing issues using the debug log, dramatically improving the debugging experience. The fixed dimensions and scrolling functionality of the details panel provide a consistent, user-friendly experience when examining long task logs.

## Usage Instructions

### Building the Application
```bash
cd /home/fl118890/Workspace/code/ansible-logs-tui
go build -o ansible-logs-view ./cmd/ansible-logs-view
```

### Running the Application
```bash
./ansible-logs-view /path/to/ansible-log-file.log
```

Or with debug enabled:
```bash
./ansible-logs-view --debug /path/to/ansible-log-file.log
```

### Keyboard Controls
- `↑` / `↓` : Navigate between tasks
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
The application creates a `debug.log` file that contains detailed information about each parsed task, including:
- Task ID
- Description
- Status
- Host
- Path
- Start time
- Diff information
- First 1000 characters of RawText

To enable debug logging, run the application with the `--debug` flag:
```bash
./ansible-logs-view --debug /path/to/ansible-log-file.log
```