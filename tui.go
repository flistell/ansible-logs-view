package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	appStyle = lipgloss.NewStyle().Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Bold(true)

	statusOkStyle = statusStyle.Copy().Background(lipgloss.Color("#25A065"))
	statusChangedStyle = statusStyle.Copy().Background(lipgloss.Color("#FFA500"))
	statusSkippingStyle = statusStyle.Copy().Background(lipgloss.Color("#888888"))
	statusFailedStyle = statusStyle.Copy().Background(lipgloss.Color("#FF0000"))
	statusUnknownStyle = statusStyle.Copy().Background(lipgloss.Color("#888888"))
	
	// Styles for tree view
	expandedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#25A065"))
	collapsedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500"))
	
	// Detail styles within tree items
	detailStyle = lipgloss.NewStyle().
			PaddingLeft(4).
			Foreground(lipgloss.Color("#AAAAAA"))
			
	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065"))
)

type model struct {
	tasks     []Task
	selected  int
	expanded  map[int]bool // Track which tasks are expanded
	width     int
	height    int
	loaded    bool
	err       error
	quitting  bool
}

func newModel(tasks []Task) model {
	return model{
		tasks:    tasks,
		selected: 0,
		expanded: make(map[int]bool),
		width:    80,
		height:   24,
		loaded:   true,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
		case "down", "j":
			if m.selected < len(m.tasks)-1 {
				m.selected++
			}
		case "enter", " ":
			// Toggle expansion of selected item
			m.expanded[m.selected] = !m.expanded[m.selected]
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	
	return m, nil
}

func (m model) View() string {
	if m.quitting {
		return ""
	}

	if !m.loaded {
		return appStyle.Render("Loading...")
	}

	if m.err != nil {
		return appStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	// Build the view
	var b strings.Builder
	
	// Title
	b.WriteString(titleStyle.Render("Ansible Tasks") + "\n\n")
	
	// Calculate visible items based on height
	visibleCount := m.height - 5 // Leave room for title and padding
	
	// Render tasks
	for i, task := range m.tasks {
		if i >= visibleCount {
			b.WriteString(fmt.Sprintf("... and %d more tasks\n", len(m.tasks)-visibleCount))
			break
		}
		
		// Format: "TASK NUMBER. TASK TITLE. TASK STATUS"
		title := fmt.Sprintf("%d. %s", task.ID, task.Description)
		status := strings.ToUpper(task.Status)
		
		// Style based on status
		var statusStyle lipgloss.Style
		switch task.Status {
		case "ok":
			statusStyle = statusOkStyle
		case "changed":
			statusStyle = statusChangedStyle
		case "skipping":
			statusStyle = statusSkippingStyle
		case "failed":
			statusStyle = statusFailedStyle
		default:
			statusStyle = statusUnknownStyle
		}
		
		statusStr := statusStyle.Render(status)
		
		// Add expansion indicator
		var indicator string
		if m.expanded[i] {
			indicator = expandedStyle.Render("▼")
		} else {
			indicator = collapsedStyle.Render("▶")
		}
		
		// Highlight if selected
		var line string
		if i == m.selected {
			line = selectedStyle.Render(fmt.Sprintf("> %s %s %s", indicator, title, statusStr))
		} else {
			line = fmt.Sprintf("  %s %s %s", indicator, title, statusStr)
		}
		
		b.WriteString(line + "\n")
		
		// If expanded, show details
		if m.expanded[i] {
			details := fmt.Sprintf("Host: %s\nPath: %s\nStart Time: %s\nStatus: %s", 
				task.Host, 
				task.Path, 
				task.StartTime.Format("2006-01-02 15:04:05"), 
				task.Status)
			
			indentedDetails := detailStyle.Render(details)
			b.WriteString(indentedDetails + "\n")
		}
	}
	
	// Instructions
	b.WriteString("\n↑/↓: Navigate • Enter: Expand/Collapse • q/Esc/Ctrl+C: Quit")
	
	return appStyle.Render(b.String())
}