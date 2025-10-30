package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
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
			
	// Diff panel styles
	diffPanelStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#25A065")).
			Padding(1, 2)
			
	diffTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1).
			MarginBottom(1)
			
	// Help text style
	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Italic(true)
)

type model struct {
	tasks         []Task
	selected      int
	expanded      map[int]bool // Track which tasks are expanded
	width         int
	height        int
	loaded        bool
	err           error
	quitting      bool
	viewport      viewport.Model
	diffViewport  viewport.Model
	offset        int // Vertical offset for viewport scrolling
}

func newModel(tasks []Task) model {
	// Create viewports
	vp := viewport.New(80, 20)
	diffVp := viewport.New(80, 10)
	
	return model{
		tasks:        tasks,
		selected:     0,
		expanded:     make(map[int]bool),
		width:        80,
		height:       24,
		loaded:       true,
		viewport:     vp,
		diffViewport: diffVp,
		offset:       0,
	}
}

func (m model) Init() tea.Cmd {
	// Initialize viewport content
	m.viewport.SetContent(m.renderTaskList())
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
				// Adjust viewport offset if needed
				if m.selected < m.offset {
					m.offset = m.selected
				}
				// Update viewport content
				m.viewport.SetContent(m.renderTaskList())
			}
		case "down", "j":
			if m.selected < len(m.tasks)-1 {
				m.selected++
				// Adjust viewport offset if needed
				visibleItems := m.viewport.Height - 2 // Account for padding
				if m.selected >= m.offset+visibleItems {
					m.offset = m.selected - visibleItems + 1
				}
				// Update viewport content
				m.viewport.SetContent(m.renderTaskList())
			}
		case "enter", " ":
			// Toggle expansion of selected item
			m.expanded[m.selected] = !m.expanded[m.selected]
			// Update viewport content
			m.viewport.SetContent(m.renderTaskList())
		case "g":
			// Go to top
			m.selected = 0
			m.offset = 0
			// Update viewport content
			m.viewport.SetContent(m.renderTaskList())
		case "G":
			// Go to bottom
			m.selected = len(m.tasks) - 1
			visibleItems := m.viewport.Height - 2 // Account for padding
			if len(m.tasks) > visibleItems {
				m.offset = len(m.tasks) - visibleItems
			} else {
				m.offset = 0
			}
			// Update viewport content
			m.viewport.SetContent(m.renderTaskList())
		}
		
		// Handle viewport key messages
		m.viewport, _ = m.viewport.Update(msg)
		m.diffViewport, _ = m.diffViewport.Update(msg)
		
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		
		// Update viewport sizes
		m.viewport.Width = m.width - 4 // Account for padding
		m.diffViewport.Width = m.width - 4
		
		// Calculate heights based on whether diff panel is visible
		if m.expanded[m.selected] && m.selected < len(m.tasks) && m.tasks[m.selected].Diff != "" {
			// Split screen: 2/3 for task list, 1/3 for diff
			m.viewport.Height = (m.height - 10) * 2 / 3
			m.diffViewport.Height = (m.height - 10) / 3
		} else {
			// Full screen for task list
			m.viewport.Height = m.height - 8
			m.diffViewport.Height = 0
		}
		
		// Update viewport content
		m.viewport.SetContent(m.renderTaskList())
		if m.selected < len(m.tasks) && m.tasks[m.selected].Diff != "" {
			m.diffViewport.SetContent(m.tasks[m.selected].Diff)
		}
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
	
	// Task list viewport
	b.WriteString(m.viewport.View() + "\n")
	
	// Show diff panel if selected task is expanded and has diff
	if m.expanded[m.selected] && m.selected < len(m.tasks) && m.tasks[m.selected].Diff != "" {
		b.WriteString("\n" + m.renderDiffPanel(m.tasks[m.selected]) + "\n")
	}
	
	// Help text
	b.WriteString("\n" + helpStyle.Render("↑/↓: Navigate • Enter: Expand/Collapse • g/G: Top/Bottom • q/Esc/Ctrl+C: Quit"))
	
	return appStyle.Render(b.String())
}

func (m model) renderTaskList() string {
	var b strings.Builder
	
	// Calculate visible range
	visibleItems := m.viewport.Height - 2 // Account for padding
	if visibleItems <= 0 {
		visibleItems = 10 // Default value
	}
	
	start := m.offset
	end := start + visibleItems
	
	if end > len(m.tasks) {
		end = len(m.tasks)
	}
	
	// Render visible tasks
	for i := start; i < end; i++ {
		task := m.tasks[i]
		
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
	
	return b.String()
}

func (m model) renderDiffPanel(task Task) string {
	if task.Diff == "" {
		return ""
	}
	
	// Create diff panel
	title := diffTitleStyle.Render(fmt.Sprintf("Diff for Task #%d: %s", task.ID, task.Description))
	
	// Update diff viewport content
	m.diffViewport.SetContent(task.Diff)
	
	content := fmt.Sprintf("%s\n%s", title, m.diffViewport.View())
	
	return diffPanelStyle.Render(content)
}