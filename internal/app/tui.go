package app

import (
	"fmt"
	"strings"
	"time"

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

	statusOkStyle       = statusStyle.Background(lipgloss.Color("#25A065"))
	statusChangedStyle  = statusStyle.Background(lipgloss.Color("#FFA500"))
	statusSkippingStyle = statusStyle.Background(lipgloss.Color("#888888"))
	statusFailedStyle   = statusStyle.Background(lipgloss.Color("#FF0000"))
	statusUnknownStyle  = statusStyle.Background(lipgloss.Color("#888888"))

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
				Padding(1, 2)

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
	StartTime   time.Time
	Status      string
	Host        string
	Path        string
	Diff        string
	RawText     string
	IsExpanded  bool
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
			StartTime:   task.StartTime,
			Status:      task.Status,
			Host:        task.Host,
			Path:        task.Path,
			Diff:        task.Diff,
			RawText:     task.RawText,
			IsExpanded:  false,
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
	nodes             []TreeNode
	filteredNodes     []TreeNode
	flatNodes         []flatNode // All visible nodes in a flat list
	selected          int
	width             int
	height            int
	loaded            bool
	err               error
	quitting          bool
	nodesViewport     viewport.Model
	detailsViewport   viewport.Model
	helpTextViewport  viewport.Model
	filterInput       textinput.Model
	showingFilter     bool
	expandedNodeCount int
	expandedNodeSize  int
	helpText          string
}

func NewModel(tasks []Task, enableDebug bool) Model {
	setupLogger(enableDebug)
	debugLog.Printf("NewModel() - Received %d tasks", len(tasks))

	nodes := convertTasksToNodes(tasks)
	debugLog.Printf("NewModel() - Converted to %d nodes", len(nodes))

	nodesVp := viewport.New(0, 0)            // Let updateViewports set the dimensions
	nodesVp.HighPerformanceRendering = false // Try without high performance mode

	detailsVp := viewport.New(0, 0)
	detailsVp.HighPerformanceRendering = false

	helpVp := viewport.New(0, 0)
	helpVp.HighPerformanceRendering = false

	ti := textinput.New()
	ti.Placeholder = "Filter..."
	ti.Prompt = "> "
	ti.CharLimit = 100
	ti.Width = 30

	m := Model{
		nodes:             nodes,
		selected:          0,
		width:             80,
		height:            24,
		loaded:            true,
		nodesViewport:     nodesVp,
		detailsViewport:   detailsVp,
		helpTextViewport:  helpVp,
		filterInput:       ti,
		helpText:          "j/k, up/down: move • ctrl+j/k: scroll details • /: filter • g/G: go to first/last line • q: quit",
		expandedNodeCount: 0,
		expandedNodeSize:  4,
	}

	// Initialize the filtered nodes and build flat nodes
	m.filteredNodes = nodes
	m.rebuildFlatNodes()

	// Update viewports to set dimensions and content
	m.updateViewports()

	debugLog.Printf("NewModel() - Initial viewport content length: %d", len(m.nodesViewport.View()))

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
				// update viewport content without full resize (reset to top)
				m.setNodeListContentFrom(strings.TrimSpace(m.renderNodeList()))
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
				debugLog.Printf("Update() - Moving up, new selected index: %d", m.selected)
				if m.selected < m.nodesViewport.YOffset {
					m.nodesViewport.SetYOffset(m.selected)
				}
				nodeList := m.renderNodeList()
				// set content but preserve the Y offset that was just adjusted above
				m.setNodeListContentPreserve(nodeList)
				m.updateDetailsViewportContent()
			}
		case "down", "j":
			if m.selected < len(m.flatNodes)-1 {
				m.selected++
				debugLog.Printf("Update() - Moving down, new selected index: %d", m.selected)
				debugLog.Printf("Update() - Before adjust: selected: %d, Y offset: %d, Height: %d", m.selected, m.nodesViewport.YOffset, m.nodesViewport.Height)
				if m.selected+(m.expandedNodeCount*m.expandedNodeSize) >= m.nodesViewport.YOffset+m.nodesViewport.Height {
					m.nodesViewport.SetYOffset(m.selected - m.nodesViewport.Height + (m.expandedNodeCount * m.expandedNodeSize) + 1)
				}
				debugLog.Printf("Update() - After adjust Y offset: %d, Height: %d", m.nodesViewport.YOffset, m.nodesViewport.Height)
				nodeList := m.renderNodeList()
				// set content but preserve the Y offset that was just adjusted above
				m.setNodeListContentPreserve(nodeList)
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
				// Ensure selected node is visible after expand/collapse
				if m.selected+(m.expandedNodeCount*m.expandedNodeSize) >= m.nodesViewport.YOffset+m.nodesViewport.Height {
					m.nodesViewport.SetYOffset(m.selected - m.nodesViewport.Height + (m.expandedNodeCount * m.expandedNodeSize) + 1)
				}
			}
		case "g":
			m.nodesViewport.GotoTop()
			m.selected = 0
			m.updateViewports()
		case "G":
			if len(m.flatNodes) > 0 {
				m.nodesViewport.GotoBottom()
				m.selected = len(m.flatNodes) - 1
				m.updateViewports()
				if m.selected+(m.expandedNodeCount*m.expandedNodeSize) >= m.nodesViewport.YOffset+m.nodesViewport.Height {
					m.nodesViewport.SetYOffset(m.selected - m.nodesViewport.Height + (m.expandedNodeCount * m.expandedNodeSize) + 1)
				}
			}
		case "pgup", "ctrl+u":
			m.detailsViewport, cmd = m.detailsViewport.Update(msg)
			cmds = append(cmds, cmd)
		case "pgdn", "ctrl+d":
			m.detailsViewport, cmd = m.detailsViewport.Update(msg)
			cmds = append(cmds, cmd)
		case "ctrl+k":
			debugLog.Printf("Update() - Scrolling details up, current yOffset: %d", m.detailsViewport.YOffset)
			m.detailsViewport.ScrollUp(1)
			debugLog.Printf("Update() - New yOffset: %d", m.detailsViewport.YOffset)
		case "ctrl+j":
			debugLog.Printf("Update() - Scrolling details down, current yOffset: %d", m.detailsViewport.YOffset)
			m.detailsViewport.ScrollDown(1)
			debugLog.Printf("Update() - New yOffset: %d", m.detailsViewport.YOffset)
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
	debugLog.Printf("rebuildFlatNodes() - Rebuilding flat nodes from %d filtered nodes", len(m.filteredNodes))
	m.flattenNodes(m.filteredNodes, 0)
	debugLog.Printf("rebuildFlatNodes() - Built %d flat nodes", len(m.flatNodes))
	// Recompute expanded node count based on the flattened nodes so the value
	// is preserved on the model (don't mutate it from render functions).
	m.expandedNodeCount = 0
	for _, fn := range m.flatNodes {
		if fn.node.IsExpanded {
			m.expandedNodeCount++
		}
	}
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
		headerHeight      = 2
		helpHeight        = 1
		detailsMinHeight  = 15
		minNodesHeight    = 3
		horizontalPadding = 4
	)

	// Calculate available space
	debugLog.Printf("updateViewports() - Calculating viewports with expandedNodeCount: %d", m.expandedNodeCount)

	// Calculate base available height
	baseHeight := m.height - headerHeight - helpHeight - 4

	// Details panel height - either minimum or 1/3 of screen (whichever is smaller)
	detailsHeight := detailsMinHeight
	if baseHeight/3 > detailsMinHeight {
		detailsHeight = baseHeight / 3
	}

	// Calculate space for nodes viewport
	nodesViewportHeight := baseHeight - detailsHeight
	if nodesViewportHeight < minNodesHeight {
		nodesViewportHeight = minNodesHeight
		// If we need to shrink details to accommodate minimum node height
		detailsHeight = baseHeight - minNodesHeight
		if detailsHeight < 3 { // Minimum for details
			detailsHeight = 3
			nodesViewportHeight = baseHeight - detailsHeight
		}
	}

	// Render node list to get content
	nodeList := strings.TrimSpace(m.renderNodeList())
	if nodeList == "" {
		nodeList = "No nodes available."
	}

	// Debug logging: viewport sizes
	debugLog.Printf("updateViewports() - Viewport sizes - total height: %d, nodesViewportHeight: %d, detailsHeight: %d",
		m.height, nodesViewportHeight, detailsHeight)

	// Assign viewport dimensions and content using helper methods
	m.assignViewportDimensions(horizontalPadding, nodesViewportHeight, detailsHeight)
	m.setNodeListContentFrom(nodeList)
	m.updateDetailsViewportContent()

	m.helpTextViewport.SetContent(m.renderHelpLine())
}

// assignViewportDimensions sets width/height on viewports and syncs input width.
func (m *Model) assignViewportDimensions(horizontalPadding, nodesViewportHeight, detailsHeight int) {
	m.nodesViewport.Width = m.width - horizontalPadding
	m.nodesViewport.Height = nodesViewportHeight
	m.detailsViewport.Width = m.width - horizontalPadding

	// Keep filter input width in sync with viewports
	if m.nodesViewport.Width >= 2 {
		m.filterInput.Width = m.nodesViewport.Width - 2
	} else {
		m.filterInput.Width = m.nodesViewport.Width
	}

	// Set details viewport height (account for title and padding)
	detailsTitleHeight := lipgloss.Height(m.renderDetailsPanelTitle())
	h := detailsHeight - detailsTitleHeight - 3
	if h < 0 {
		h = 0
	}
	m.detailsViewport.Height = h
}

// setNodeListContentFrom sets the rendered node list into the nodes viewport
// and ensures the viewport offset is in a sane state.
func (m *Model) setNodeListContentFrom(nodeList string) {
	if strings.TrimSpace(nodeList) == "" {
		nodeList = "No nodes available."
	}
	debugLog.Printf("setNodeListContentFrom() - Node list content length: %d", len(nodeList))
	m.nodesViewport.SetContent(nodeList)
	debugLog.Printf("setNodeListContentFrom() - Viewport dimensions: w=%d h=%d", m.nodesViewport.Width, m.nodesViewport.Height)
	m.nodesViewport.GotoTop()
}

// setNodeListContentPreserve sets node list content but preserves current Y offset
// (useful when callers adjust YOffset before updating content).
func (m *Model) setNodeListContentPreserve(nodeList string) {
	if strings.TrimSpace(nodeList) == "" {
		nodeList = "No nodes available."
	}
	// capture current offset (may have been adjusted by caller)
	cur := m.nodesViewport.YOffset
	m.nodesViewport.SetContent(nodeList)
	// Ensure offset is within valid range given new content height
	maxOffset := 0
	lines := strings.Count(nodeList, "\n") + 1
	if lines > m.nodesViewport.Height {
		maxOffset = lines - m.nodesViewport.Height
	}
	if cur > maxOffset {
		cur = maxOffset
	}
	if cur < 0 {
		cur = 0
	}
	m.nodesViewport.SetYOffset(cur)
}

func (m *Model) updateDetailsViewportContent() {
	if len(m.flatNodes) == 0 || m.selected < 0 || m.selected >= len(m.flatNodes) {
		m.detailsViewport.SetContent("No node selected.")
		return
	}
	selectedNode := m.flatNodes[m.selected].node

	// Create content with title
	replacer := strings.NewReplacer("\\n", "\n", "\\t", "\t", "\\\"", "\"")
	detailsContent := fmt.Sprintf("Item: %s\n\n%s",
		selectedNode.Name,
		replacer.Replace(selectedNode.Description))

	// Calculate the available width for content, accounting for borders and padding
	contentWidth := m.detailsViewport.Width - 4 // -4 for left and right padding/borders

	// Style the content with fixed width to enable proper scrolling
	styledContent := lipgloss.NewStyle().
		Width(contentWidth).
		Render(detailsContent)

	debugLog.Printf("updateDetailsViewportContent() - Details content length: %d lines", strings.Count(styledContent, "\n")+1)

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
	var mainSections []string
	if m.showingFilter {
		// show filter input above the node list
		mainSections = append(mainSections, m.filterInput.View())
	}

	// Add nodes viewport
	mainSections = append(mainSections, m.nodesViewport.View())

	// Add details panel and help text as a separate section to anchor to bottom
	bottomSection := lipgloss.JoinVertical(lipgloss.Left,
		m.renderDetailsPanel(),
		m.renderHelpLine(),
	)

	// Calculate how much vertical space is available for the main content
	// after header and padding are accounted for
	headerHeight := lipgloss.Height(header)

	// Calculate padding height (appStyle includes padding)
	availableHeight := m.height - headerHeight - 4 // account for appStyle padding

	// Calculate the height of the bottom section
	bottomHeight := lipgloss.Height(bottomSection)

	// Calculate the height available for the main sections
	var mainSectionsContent string
	if len(mainSections) > 0 {
		// Join the main sections together
		mainContent := lipgloss.JoinVertical(lipgloss.Left, mainSections...)
		mainContentHeight := lipgloss.Height(mainContent)

		// If the main content is smaller than available space,
		// we need to add padding to push the bottom section down
		if mainContentHeight+bottomHeight < availableHeight {
			// Expand main content to fill available space
			expandedContentHeight := availableHeight - bottomHeight
			// Ensure minimum height for main content
			if expandedContentHeight < 3 {
				expandedContentHeight = 3
			}

			mainContentStyle := lipgloss.NewStyle().Height(expandedContentHeight).MaxHeight(expandedContentHeight)
			mainSectionsContent = mainContentStyle.Render(mainContent)
		} else {
			mainSectionsContent = mainContent
		}
	} else {
		// If there's no main content, just render the bottom section
		mainSectionsContent = ""
	}

	// Combine main content with the bottom section
	fullMainContent := lipgloss.JoinVertical(lipgloss.Left,
		mainSectionsContent,
		bottomSection,
	)

	// Join header with padded content
	finalView := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		appStyle.Render(fullMainContent),
	)

	return finalView
}

func (m *Model) renderHelpLine() string {
	debugLog.Printf("renderHelpLine() - Rendering help line with %d expanded nodes", m.expandedNodeCount)
	viewportContent := m.helpTextViewport.View()
	contentText := fmt.Sprintf("%s", m.helpText)
	content := lipgloss.JoinVertical(lipgloss.Left, contentText, viewportContent)
	return helpStyle.Width(m.width - 4).Render(content)
}

func (m Model) renderNodeList() string {
	var b strings.Builder
	debugLog.Printf("renderNodeList() - Rendering %d nodes, selected index: %d", len(m.flatNodes), m.selected)
	// Use a local counter when rendering so we don't mutate model state here.
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
		case "failed", "fatal":
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
			debugLog.Printf("renderNodeList() - Highlighting line %d: %s", i, line)
			selectedLineStyle := selectedStyle.Copy().Width(m.width - 4)
			line = selectedLineStyle.Render(line)
		}
		b.WriteString(line + "\n")
		// If the node is expanded, show its description as an indented detail
		if node.IsExpanded && strings.TrimSpace(node.Description) != "" {
			descLine := fmt.Sprintf("Host: %s\nPath: %s\nStart Time: %s\nStatus: %s",
				node.Host,
				node.Path,
				node.StartTime.Format("2006-01-02 15:04:05"),
				node.Status)

			b.WriteString(inlineDetailStyle.Render(descLine) + "\n")
		}
	}
	content := b.String()
	if len(content) > 0 {
		debugLog.Printf("renderNodeList() - First line of content: %s", strings.Split(content, "\n")[0])
	}
	debugLog.Printf("renderNodeList() - Computed expanded nodes in model %d", m.expandedNodeCount)
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
	term = strings.ToLower(term)
	if term == "" {
		m.filteredNodes = m.nodes
	} else {
		var filtered []TreeNode
		for _, n := range m.nodes {
			// Check against all possible fields
			if strings.Contains(strings.ToLower(n.Name), term) ||
				strings.Contains(strings.ToLower(n.Status), term) ||
				strings.Contains(strings.ToLower(n.Host), term) ||
				strings.Contains(strings.ToLower(n.Path), term) ||
				strings.Contains(n.StartTime.Format("2006-01-02 15:04:05"), term) ||
				strings.Contains(n.StartTime.Format("2006-01-02"), term) ||
				strings.Contains(n.StartTime.Format("15:04:05"), term) {
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

func (m *Model) applyFuzzyFilter(term string) {
	term = strings.TrimSpace(term)
	if term == "" {
		m.filteredNodes = m.nodes
	} else {
		var filtered []TreeNode
		for _, n := range m.nodes {
			if fuzzyMatch(term, n.Name) ||
				fuzzyMatch(term, n.Status) ||
				fuzzyMatch(term, n.Host) ||
				fuzzyMatch(term, n.Path) ||
				fuzzyMatch(term, n.StartTime.Format("2006-01-02 15:04:05")) {
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
