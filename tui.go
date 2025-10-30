package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
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
			
	// Details panel styles
	detailsPanelStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#25A065")).
			Padding(1, 2)
			
	detailsTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1).
			MarginBottom(1)
			
	// Help text style
	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Italic(true)
			
	// Filter input style
	filterStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#888888")).
			Padding(0, 1)
)

type model struct {
	tasks         []Task
	filteredTasks []Task
	selected      int
	expanded      map[int]bool // Track which tasks are expanded
	width         int
	height        int
	loaded        bool
	err           error
	quitting      bool
	viewport      viewport.Model
	detailsViewport  viewport.Model
	offset        int // Vertical offset for viewport scrolling
	filterInput   textinput.Model
	showingFilter bool
	showingDetails bool // Whether to show details panel
}

func newModel(tasks []Task) model {
	// Create viewports
	vp := viewport.New(80, 20)
	detailsVp := viewport.New(80, 10)
	
	// Create text input for filtering
	ti := textinput.New()
	ti.Placeholder = "Filter tasks..."
	ti.Prompt = "🔍 "
	ti.CharLimit = 100
	ti.Width = 30
	
	return model{
		tasks:         tasks,
		filteredTasks: tasks, // Initially, no filter
		selected:      0,
		expanded:      make(map[int]bool),
		width:         80,
		height:        24,
		loaded:        true,
		viewport:      vp,
		detailsViewport: detailsVp,
		offset:        0,
		filterInput:   ti,
		showingFilter: false,
		showingDetails: false,
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
		// Handle filter input first if it's focused
		if m.showingFilter {
			var cmd tea.Cmd
			m.filterInput, cmd = m.filterInput.Update(msg)
			
			switch msg.String() {
			case "esc":
				// Cancel filter and restore all tasks
				m.showingFilter = false
				m.filterInput.Blur()
				m.filterInput.SetValue("")
				m.filteredTasks = m.tasks
				m.selected = 0
				m.offset = 0
				// Update viewport content
				m.viewport.SetContent(m.renderTaskList())
				return m, nil
			case "enter":
				// Apply filter
				m.showingFilter = false
				m.filterInput.Blur()
				m.applyFilter(m.filterInput.Value())
				m.selected = 0
				m.offset = 0
				// Update viewport content
				m.viewport.SetContent(m.renderTaskList())
				return m, nil
			}
			
			return m, cmd
		}
		
		// Handle navigation and other keys
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "/":
			// Toggle filter input
			m.showingFilter = !m.showingFilter
			if m.showingFilter {
				m.filterInput.Focus()
				return m, textinput.Blink
			} else {
				m.filterInput.Blur()
				return m, nil
			}
		case "esc":
			if m.showingFilter {
				// Cancel filter and restore all tasks
				m.showingFilter = false
				m.filterInput.Blur()
				m.filterInput.SetValue("")
				m.filteredTasks = m.tasks
				m.selected = 0
				m.offset = 0
				// Update viewport content
				m.viewport.SetContent(m.renderTaskList())
				return m, nil
			}
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
			if m.selected < len(m.filteredTasks)-1 {
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
			if len(m.filteredTasks) > 0 {
				// Toggle expansion of selected item
				// Find the index of the selected task in the original tasks list
				originalIndex := -1
				for i, originalTask := range m.tasks {
					if originalTask.ID == m.filteredTasks[m.selected].ID {
						originalIndex = i
						break
					}
				}
				
				if originalIndex != -1 {
					// Toggle expansion state
					m.expanded[originalIndex] = !m.expanded[originalIndex]
					
					// If expanded, show the full raw task text in the details panel
					if m.expanded[originalIndex] {
						m.showingDetails = true
						// Set the raw task text as the content for the details viewport
						taskRawText := m.filteredTasks[m.selected].RawText
						m.detailsViewport.SetContent(taskRawText)
					} else {
						m.showingDetails = false
					}
				}
				
				// Update viewport content
				m.viewport.SetContent(m.renderTaskList())
			}
		case "g":
			// Go to top
			m.selected = 0
			m.offset = 0
			// Update viewport content
			m.viewport.SetContent(m.renderTaskList())
		case "G":
			// Go to bottom
			m.selected = len(m.filteredTasks) - 1
			visibleItems := m.viewport.Height - 2 // Account for padding
			if len(m.filteredTasks) > visibleItems {
				m.offset = len(m.filteredTasks) - visibleItems
			} else {
				m.offset = 0
			}
			// Update viewport content
			m.viewport.SetContent(m.renderTaskList())
		}
		
		// Handle viewport key messages
		m.viewport, _ = m.viewport.Update(msg)
		m.detailsViewport, _ = m.detailsViewport.Update(msg)
		
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		
		// Update viewport sizes
		m.viewport.Width = m.width - 4 // Account for padding
		m.detailsViewport.Width = m.width - 4
		
		// Calculate heights based on whether details panel is visible
		if m.showingFilter || (len(m.filteredTasks) > 0 && m.showingDetails) {
			// Split screen: 2/3 for task list, 1/3 for details or filter input
			m.viewport.Height = (m.height - 10) * 2 / 3
			m.detailsViewport.Height = (m.height - 10) / 3
		} else {
			// Full screen for task list
			m.viewport.Height = m.height - 8
			m.detailsViewport.Height = 0
		}
		
		// Update viewport content
		m.viewport.SetContent(m.renderTaskList())
		if len(m.filteredTasks) > 0 && m.showingDetails {
			taskRawText := m.filteredTasks[m.selected].RawText
			m.detailsViewport.SetContent(taskRawText)
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
	
	// Title with filter input
	title := titleStyle.Render("Ansible Tasks")
	if m.showingFilter {
		b.WriteString(fmt.Sprintf("%s %s\n\n", title, m.filterInput.View()))
	} else {
		b.WriteString(fmt.Sprintf("%s\n\n", title))
	}
	
	// Task list viewport
	b.WriteString(m.viewport.View() + "\n")
	
	// Show details panel if selected task is expanded
	if len(m.filteredTasks) > 0 && m.showingDetails {
		b.WriteString("\n" + m.renderDetailsPanel(m.filteredTasks[m.selected]) + "\n")
	}
	
	// Help text
	helpText := "↑/↓: Navigate • Enter: Expand/Collapse • g/G: Top/Bottom • /: Filter • q/Ctrl+C: Quit"
	if m.showingFilter {
		helpText = "/: Filter • Esc: Cancel • Enter: Apply"
	}
	b.WriteString("\n" + helpStyle.Render(helpText))
	
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
	
	if end > len(m.filteredTasks) {
		end = len(m.filteredTasks)
	}
	
	// Render visible tasks
	for i := start; i < end; i++ {
		task := m.filteredTasks[i]
		
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
		taskIndex := m.getIndexInOriginalTaskList(i)
		if taskIndex != -1 && m.expanded[taskIndex] {
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
		
		// If expanded, show details in the inline view
		if taskIndex != -1 && m.expanded[taskIndex] {
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

func (m model) renderDetailsPanel(task Task) string {
	// Create details panel
	title := detailsTitleStyle.Render(fmt.Sprintf("Details for Task #%d: %s", task.ID, task.Description))
	
	// Set the raw task text as the content for the details viewport
	m.detailsViewport.SetContent(task.RawText)
	
	// Get the rendered content from the viewport
	viewportContent := m.detailsViewport.View()
	
	// Combine the title and viewport content
	content := fmt.Sprintf("%s\n%s", title, viewportContent)
	
	return detailsPanelStyle.Render(content)
}

// applyFilter filters tasks based on the provided search term
func (m *model) applyFilter(term string) {
	term = strings.ToLower(term)
	if term == "" {
		m.filteredTasks = m.tasks
		return
	}
	
	var filtered []Task
	for _, task := range m.tasks {
		// Check against all possible fields
		if strings.Contains(strings.ToLower(task.Description), term) ||
			strings.Contains(strings.ToLower(task.Status), term) ||
			strings.Contains(strings.ToLower(task.Host), term) ||
			strings.Contains(strings.ToLower(task.Path), term) ||
			strings.Contains(task.StartTime.Format("2006-01-02 15:04:05"), term) ||
			strings.Contains(task.StartTime.Format("2006-01-02"), term) ||
			strings.Contains(task.StartTime.Format("15:04:05"), term) ||
			strings.Contains(strings.ToLower(task.Diff), term) ||
			strings.Contains(strings.ToLower(task.RawText), term) {
			filtered = append(filtered, task)
		}
	}
	
	m.filteredTasks = filtered
}

// getIndexInOriginalTaskList gets the index of the filtered task in the original tasks list
func (m *model) getIndexInOriginalTaskList(filteredIndex int) int {
	if filteredIndex < 0 || filteredIndex >= len(m.filteredTasks) {
		return -1
	}
	
	task := m.filteredTasks[filteredIndex]
	
	for i, originalTask := range m.tasks {
		if originalTask.ID == task.ID {
			return i
		}
	}
	
	return -1
}