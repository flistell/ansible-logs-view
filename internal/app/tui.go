package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// package-level logger is provided from logger.go



var (
	appStyle = lipgloss.NewStyle().Padding(1, 2)

	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Bold(true).
			Padding(0, 2).
			MarginBottom(1)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Bold(true)

	statusOkStyle       = statusStyle.Copy().Background(lipgloss.Color("#25A065"))
	statusChangedStyle  = statusStyle.Copy().Background(lipgloss.Color("#FFA500"))
	statusSkippingStyle = statusStyle.Copy().Background(lipgloss.Color("#888888"))
	statusFailedStyle   = statusStyle.Copy().Background(lipgloss.Color("#FF0000"))
	statusUnknownStyle  = statusStyle.Copy().Background(lipgloss.Color("#888888"))

	selectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#25A065")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true)

	// Inline detail style for expanded nodes
	inlineDetailStyle = lipgloss.NewStyle().
				PaddingLeft(4).
				Foreground(lipgloss.Color("#AAAAAA"))

	// Details panel styles
	detailsPanelStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("#25A065")).
				Padding(1, 2).
				Height(10) // Fixed height for details panel

	detailsTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFDF5")).
				Background(lipgloss.Color("#25A065")).
				Padding(0, 1)

	// Help text style
	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Italic(true)
)

// TreeNode represents a node in our tree structure
type TreeNode struct {
	ID          int
	Name        string
	Description string
	IsExpanded  bool
	Status	  string
}

// flatNode represents a node in the flattened tree for display
type flatNode struct {
	node  *TreeNode
	depth int
}



// Convert tasks to tree nodes
func convertTasksToNodes(tasks []Task) []TreeNode {
	nodes := make([]TreeNode, len(tasks))
	for i, task := range tasks {
		nodes[i] = TreeNode{
			ID:          task.ID,
			Name:        task.Description,
			Description: task.RawText,
			IsExpanded:  false,
			Status: task.Status,
		}
	}
	return nodes
}

// fuzzyMatch performs a simple fuzzy match: all characters in pattern must
// appear in order in s (case-insensitive). This is cheap and good for
// interactive filtering.
func fuzzyMatch(pattern, s string) bool {
	pattern = strings.ToLower(pattern)
	s = strings.ToLower(s)
	if pattern == "" {
		return true
	}
	si := 0
	for _, pr := range pattern {
		idx := strings.IndexRune(s[si:], pr)
		if idx < 0 {
			return false
		}
		si += idx + 1
		if si >= len(s) && pr != rune(pattern[len(pattern)-1]) {
			// if we've reached the end of s but there are still pattern runes
			// left (and the last rune wasn't matched), it's a fail
			// (the normal IndexRune check above handles most cases)
		}
	}
	return true
}
// Model represents the TUI state (PoC)
type Model struct {
	nodes           []TreeNode
	filteredNodes   []TreeNode
	flatNodes       []flatNode // All visible nodes in a flat list
	selected        int
	width           int
	height          int
	loaded          bool
	err             error
	quitting        bool
	nodesViewport   viewport.Model
	detailsViewport viewport.Model
	filterInput     textinput.Model
	showingFilter   bool
}

func NewModel(tasks []Task, enableDebug bool) Model {
	setupLogger(enableDebug)
	debugLog.Printf("Received %d tasks", len(tasks))

	nodes := convertTasksToNodes(tasks)
	debugLog.Printf("Converted to %d nodes", len(nodes))

	nodesVp := viewport.New(0, 0) // Let updateViewports set the dimensions
	nodesVp.HighPerformanceRendering = false // Try without high performance mode

	detailsVp := viewport.New(0, 0)
	detailsVp.HighPerformanceRendering = false

	ti := textinput.New()
	ti.Placeholder = "Filter..."
	ti.Prompt = "> "
	ti.CharLimit = 100
	ti.Width = 30

	m := Model{
		nodes:           nodes,
		selected:        0,
		width:           80,
		height:          24,
		loaded:          true,
		nodesViewport:   nodesVp,
		detailsViewport: detailsVp,
		filterInput:     ti,
	}
	
	// Initialize the filtered nodes and build flat nodes
	m.filteredNodes = nodes
	m.rebuildFlatNodes()
	
	// Update viewports to set dimensions and content
	m.updateViewports()
	
	debugLog.Printf("Initial viewport content length: %d", len(m.nodesViewport.View()))
	
	return m
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateViewports()
		return m, nil

	case tea.KeyMsg:
		if m.showingFilter {
			switch msg.String() {
			case "esc":
				m.showingFilter = false
				m.filterInput.Blur()
				m.filterInput.SetValue("")
				m.applyFilter("")
				m.updateViewports()
				return m, nil
			case "enter":
				m.showingFilter = false
				m.filterInput.Blur()
				// apply final filter and close input
				m.applyFilter(m.filterInput.Value())
				m.updateViewports()
				return m, nil
			default:
				// update the input model first
				m.filterInput, cmd = m.filterInput.Update(msg)
				// apply filter as-you-type
				m.applyFilter(m.filterInput.Value())
				// update viewport content without full resize
				m.nodesViewport.SetContent(strings.TrimSpace(m.renderNodeList()))
				return m, cmd
			}
		}

		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "/":
			m.showingFilter = true
			m.filterInput.Focus()
			return m, textinput.Blink
		case "up", "k":
			if m.selected > 0 {
				m.selected--
				debugLog.Printf("Moving up, new selected index: %d", m.selected)
				if m.selected < m.nodesViewport.YOffset {
					m.nodesViewport.SetYOffset(m.selected)
				}
				nodeList := m.renderNodeList()
				m.nodesViewport.SetContent(nodeList)
				m.updateDetailsViewportContent()
			}
		case "down", "j":
			if m.selected < len(m.flatNodes)-1 {
				m.selected++
				debugLog.Printf("Moving down, new selected index: %d", m.selected)
				if m.selected >= m.nodesViewport.YOffset+m.nodesViewport.Height {
					m.nodesViewport.SetYOffset(m.selected - m.nodesViewport.Height + 1)
				}
				nodeList := m.renderNodeList()
				m.nodesViewport.SetContent(nodeList)
				m.updateDetailsViewportContent()
			}
		case "enter", "return", " ":
			if len(m.flatNodes) > 0 {
				node := m.flatNodes[m.selected].node
				node.IsExpanded = !node.IsExpanded
				oldSelected := m.selected
				m.rebuildFlatNodes()
				m.updateViewports()
				m.selected = oldSelected
			}
		case "g":
			m.selected = 0
			m.nodesViewport.GotoTop()
			m.updateDetailsViewportContent()
		case "G":
			if len(m.flatNodes) > 0 {
				m.selected = len(m.flatNodes) - 1
				m.nodesViewport.GotoBottom()
				m.updateDetailsViewportContent()
			}
		case "pgup", "ctrl+u":
			m.detailsViewport, cmd = m.detailsViewport.Update(msg)
			cmds = append(cmds, cmd)
		case "pgdn", "ctrl+d":
			m.detailsViewport, cmd = m.detailsViewport.Update(msg)
			cmds = append(cmds, cmd)
		case "ctrl+k":
			debugLog.Printf("Scrolling details up, current yOffset: %d", m.detailsViewport.YOffset)
			m.detailsViewport.LineUp(1)
			debugLog.Printf("New yOffset: %d", m.detailsViewport.YOffset)
		case "ctrl+j":
			debugLog.Printf("Scrolling details down, current yOffset: %d", m.detailsViewport.YOffset)
			m.detailsViewport.LineDown(1)
			debugLog.Printf("New yOffset: %d", m.detailsViewport.YOffset)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) flattenNodes(nodes []TreeNode, depth int) {
	for i := range nodes {
		node := &nodes[i]
		m.flatNodes = append(m.flatNodes, flatNode{node: node, depth: depth})
	}
}

func (m *Model) rebuildFlatNodes() {
	m.flatNodes = []flatNode{}
	debugLog.Printf("Rebuilding flat nodes from %d filtered nodes", len(m.filteredNodes))
	m.flattenNodes(m.filteredNodes, 0)
	debugLog.Printf("Built %d flat nodes", len(m.flatNodes))
	if m.selected >= len(m.flatNodes) {
		m.selected = len(m.flatNodes) - 1
	}
	if m.selected < 0 {
		m.selected = 0
	}
}

func (m *Model) updateViewports() {
	// Fixed sizes
	const (
		headerHeight    = 2  // Fixed header height (1 for content + 1 for margin)
		helpHeight      = 1  // Fixed help section height
		detailsHeight   = 15 // Fixed details panel height (including title and borders)
		minNodesHeight  = 3  // Minimum height for nodes viewport
		horizontalPadding = 4 // Padding for viewports (2 on each side)
	)

	// Calculate available space
	remainingHeight := m.height - headerHeight - helpHeight - detailsHeight - 4 // -4 for padding and margins
	if remainingHeight < minNodesHeight {
		remainingHeight = minNodesHeight
	}

	// Set viewport dimensions
	m.nodesViewport.Width = m.width - horizontalPadding
	m.detailsViewport.Width = m.width - horizontalPadding

	// Keep filter input width in sync with viewports so it renders full width
	m.filterInput.Width = m.nodesViewport.Width - 2

	// Set heights
	m.nodesViewport.Height = remainingHeight
	detailsTitleHeight := lipgloss.Height(m.renderDetailsPanelTitle())
	m.detailsViewport.Height = detailsHeight - detailsTitleHeight - 3 // -3 for borders and padding

	// Ensure node list content is set and viewport offset is valid
	nodeList := strings.TrimSpace(m.renderNodeList())
	if nodeList == "" {
		nodeList = "No nodes available."
	}
	debugLog.Printf("Node list content length: %d", len(nodeList))
	m.nodesViewport.SetContent(nodeList)
	debugLog.Printf("Viewport dimensions: w=%d h=%d", m.nodesViewport.Width, m.nodesViewport.Height)
	// make sure the viewport shows from top by default
	m.nodesViewport.GotoTop()

	m.updateDetailsViewportContent()
}

func (m *Model) updateDetailsViewportContent() {
	if len(m.flatNodes) == 0 || m.selected < 0 || m.selected >= len(m.flatNodes) {
		m.detailsViewport.SetContent("No node selected.")
		return
	}
	selectedNode := m.flatNodes[m.selected].node
	
	// Create content with title
	detailsContent := fmt.Sprintf("Item: %s\n\n%s",
		selectedNode.Name,
		selectedNode.Description)
		
	// Calculate the available width for content, accounting for borders and padding
	contentWidth := m.detailsViewport.Width - 4 // -4 for left and right padding/borders
	
	// Style the content with fixed width to enable proper scrolling
	styledContent := lipgloss.NewStyle().
		Width(contentWidth).
		Render(detailsContent)
		
	debugLog.Printf("Details content length: %d lines", strings.Count(styledContent, "\n")+1)
	
	m.detailsViewport.SetContent(styledContent)
	// Preserve scroll position unless selected item changed
	currentYOffset := m.detailsViewport.YOffset
	if currentYOffset == 0 {
		m.detailsViewport.GotoTop()
	}
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}

	// Header - fixed at top, full width
	header := headerStyle.
		Width(m.width).
		Render("Ansible Logs TUI")

	// Build main content area: optional filter input, nodes viewport, details panel, help
	var sections []string
	if m.showingFilter {
		// show filter input above the node list
		sections = append(sections, m.filterInput.View())
	}
	sections = append(sections, m.nodesViewport.View())
	sections = append(sections, m.renderDetailsPanel())
	sections = append(sections, helpStyle.Width(m.width-4).Render("j/k, up/down: move • ctrl+j/k: scroll details • q: quit"))

	mainContent := lipgloss.JoinVertical(lipgloss.Left, sections...)

	// Join header with padded content
	finalView := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		appStyle.Render(mainContent),
	)

	return finalView
}

func (m Model) renderNodeList() string {
	var b strings.Builder
	debugLog.Printf("Rendering %d nodes, selected index: %d", len(m.flatNodes), m.selected)
	for i, flatNode := range m.flatNodes {
		node := flatNode.node
		indent := strings.Repeat("  ", flatNode.depth)

		status := strings.ToUpper(node.Status)

		// Style based on status
		var statusStyle lipgloss.Style
		switch node.Status {
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

		indicator := " "
		if node.IsExpanded {
			indicator = "▼"
		} else {
			indicator = "▶"
		}
		line := fmt.Sprintf("%s%s [%d] %s - [%s]", indent, indicator, node.ID, node.Name, statusStr)
		if i == m.selected {
			debugLog.Printf("Highlighting line %d: %s", i, line)
			selectedLineStyle := selectedStyle.Copy().Width(m.width - 4)
			line = selectedLineStyle.Render(line)
		}
		b.WriteString(line + "\n")
		// If the node is expanded, show its description as an indented detail
		if node.IsExpanded && strings.TrimSpace(node.Description) != "" {
			descLine := fmt.Sprintf("%s  %s", indent, node.Description)
			b.WriteString(inlineDetailStyle.Render(descLine) + "\n")
		}
	}
	content := b.String()
	if len(content) > 0 {
		debugLog.Printf("First line of content: %s", strings.Split(content, "\n")[0])
	}
	return content
}

func (m Model) renderDetailsPanelTitle() string {
	return detailsTitleStyle.Render("Details")
}

func (m Model) renderDetailsPanel() string {
	title := m.renderDetailsPanelTitle()
	viewportContent := m.detailsViewport.View()
	panelContent := lipgloss.JoinVertical(lipgloss.Left, title, viewportContent)
	return detailsPanelStyle.Width(m.width - 4).Render(panelContent)
}

func (m *Model) applyFilter(term string) {
	term = strings.TrimSpace(term)
	if term == "" {
		m.filteredNodes = m.nodes
	} else {
		var filtered []TreeNode
		for _, n := range m.nodes {
			if fuzzyMatch(term, n.Name) || fuzzyMatch(term, n.Description) {
				filtered = append(filtered, n)
			}
		}
		m.filteredNodes = filtered
	}
	m.rebuildFlatNodes()

	// Reset selection and viewport to top when applying a filter
	if len(m.flatNodes) == 0 {
		m.selected = 0
	} else if m.selected >= len(m.flatNodes) {
		m.selected = len(m.flatNodes) - 1
	}
	m.nodesViewport.SetContent(strings.TrimSpace(m.renderNodeList()))
	m.nodesViewport.GotoTop()
}


