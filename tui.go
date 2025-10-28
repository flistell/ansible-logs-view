package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
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
	
	detailTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFDF5")).
				Background(lipgloss.Color("#25A065")).
				Padding(0, 1).
				MarginBottom(1)
				
	detailStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#25A065")).
			Padding(1, 2)
)

type TaskItem struct {
	task Task
}

func (t TaskItem) FilterValue() string {
	return t.task.Description
}

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(TaskItem)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", i.task.ID, i.task.Description)
	
	// Truncate the string to fit the width
	width := lipgloss.Width(str)
	if width > m.Width()-10 {
		runes := []rune(str)
		if len(runes) > m.Width()-13 {
			str = string(runes[:m.Width()-13]) + "..."
		}
	}

	// Style based on status
	var statusStyle lipgloss.Style
	switch i.task.Status {
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

	// Highlight if selected
	if index == m.Index() {
		str = "> " + str
	} else {
		str = "  " + str
	}

	// Add status indicator
	statusStr := statusStyle.Render(strings.ToUpper(i.task.Status))

	fmt.Fprintf(w, "%s %s", str, statusStr)
}

type viewState int

const (
	listView viewState = iota
	detailView
)

type model struct {
	list      list.Model
	tasks     []Task
	loaded    bool
	err       error
	quitting  bool
	viewState viewState
	selected  int
}

func newModel(tasks []Task) model {
	// Convert tasks to TaskItems
	items := make([]list.Item, len(tasks))
	for i, task := range tasks {
		items[i] = TaskItem{task: task}
	}

	// Create list
	l := list.New(items, itemDelegate{}, 0, 0)
	l.Title = "Ansible Tasks"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle

	return model{
		list:      l,
		tasks:     tasks,
		loaded:    true,
		viewState: listView,
		selected:  0,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.viewState {
	case listView:
		return m.updateListView(msg)
	case detailView:
		return m.updateDetailView(msg)
	}
	return m, nil
}

func (m model) updateListView(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			m.viewState = detailView
			m.selected = m.list.Index()
			return m, nil
		}
	case tea.WindowSizeMsg:
		h, v := appStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) updateDetailView(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "enter", "backspace":
			m.viewState = listView
			return m, nil
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.quitting {
		return ""
	}

	switch m.viewState {
	case listView:
		return m.listView()
	case detailView:
		return m.detailView()
	}
	return ""
}

func (m model) listView() string {
	if !m.loaded {
		return appStyle.Render("Loading...")
	}

	if m.err != nil {
		return appStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	return appStyle.Render(m.list.View())
}

func (m model) detailView() string {
	if m.selected >= len(m.tasks) {
		return appStyle.Render("Invalid selection")
	}
	
	task := m.tasks[m.selected]
	
	// Create detail view
	title := detailTitleStyle.Render(fmt.Sprintf("Task #%d: %s", task.ID, task.Description))
	status := fmt.Sprintf("Status: %s", task.Status)
	host := fmt.Sprintf("Host: %s", task.Host)
	path := fmt.Sprintf("Path: %s", task.Path)
	startTime := fmt.Sprintf("Start Time: %s", task.StartTime.Format("2006-01-02 15:04:05"))
	
	content := fmt.Sprintf("%s\n%s\n%s\n%s\n%s", title, status, host, path, startTime)
	
	return appStyle.Render(detailStyle.Render(content) + "\n\nPress Enter or Backspace to return to list")
}