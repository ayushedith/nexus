package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nexusapi/nexus/pkg/collection"
)

type Model struct {
	width          int
	height         int
	activePane     pane
	requestList    list.Model
	requestEditor  textarea.Model
	responseView   viewport.Model
	collection     *collection.Collection
	results        []collection.ExecutionResult
	selectedIdx    int
	keys           keyMap
	showHelp       bool
	environment    string
}

type pane int

const (
	paneList pane = iota
	paneRequest
	paneResponse
)

type keyMap struct {
	Quit      key.Binding
	NextPane  key.Binding
	PrevPane  key.Binding
	Execute   key.Binding
	Help      key.Binding
	Up        key.Binding
	Down      key.Binding
}

func defaultKeyMap() keyMap {
	return keyMap{
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c", "q"),
			key.WithHelp("q", "quit"),
		),
		NextPane: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next pane"),
		),
		PrevPane: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "prev pane"),
		),
		Execute: key.NewBinding(
			key.WithKeys("ctrl+e"),
			key.WithHelp("ctrl+e", "execute"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Up: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("↓/j", "down"),
		),
	}
}

func NewModel(coll *collection.Collection, env string) Model {
	items := []list.Item{}
	for _, req := range coll.Requests {
		items = append(items, requestItem{req})
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Requests"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)

	ta := textarea.New()
	ta.Placeholder = "Request details..."
	ta.Focus()

	vp := viewport.New(0, 0)

	return Model{
		requestList:   l,
		requestEditor: ta,
		responseView:  vp,
		collection:    coll,
		results:       []collection.ExecutionResult{},
		keys:          defaultKeyMap(),
		activePane:    paneList,
		environment:   env,
	}
}

type requestItem struct {
	req collection.Request
}

func (i requestItem) FilterValue() string { return i.req.Name }
func (i requestItem) Title() string       { return i.req.Name }
func (i requestItem) Description() string { return fmt.Sprintf("%s %s", i.req.Method, i.req.URL) }

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateSizes()
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keys.NextPane):
			m.activePane = (m.activePane + 1) % 3
			return m, nil

		case key.Matches(msg, m.keys.PrevPane):
			if m.activePane == 0 {
				m.activePane = 2
			} else {
				m.activePane--
			}
			return m, nil

		case key.Matches(msg, m.keys.Execute):
			if m.activePane == paneRequest && m.selectedIdx >= 0 {
				return m, m.executeRequest()
			}

		case key.Matches(msg, m.keys.Help):
			m.showHelp = !m.showHelp
			return m, nil
		}
	}

	var cmd tea.Cmd
	switch m.activePane {
	case paneList:
		m.requestList, cmd = m.requestList.Update(msg)
		if m.requestList.Index() != m.selectedIdx {
			m.selectedIdx = m.requestList.Index()
			m.updateRequestEditor()
		}
	case paneRequest:
		m.requestEditor, cmd = m.requestEditor.Update(msg)
	case paneResponse:
		m.responseView, cmd = m.responseView.Update(msg)
	}

	return m, cmd
}

func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	listView := m.renderList()
	requestView := m.renderRequest()
	responseView := m.renderResponse()

	leftPane := lipgloss.JoinVertical(lipgloss.Left, listView)
	rightPane := lipgloss.JoinVertical(lipgloss.Left, requestView, responseView)

	content := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)

	help := m.renderHelp()
	return lipgloss.JoinVertical(lipgloss.Left, content, help)
}

func (m *Model) updateSizes() {
	listWidth := m.width / 3
	rightWidth := m.width - listWidth - 2

	listHeight := m.height - 4
	requestHeight := (m.height - 4) / 2
	responseHeight := m.height - requestHeight - 4

	m.requestList.SetSize(listWidth, listHeight)
	m.requestEditor.SetWidth(rightWidth)
	m.requestEditor.SetHeight(requestHeight - 2)
	m.responseView.Width = rightWidth
	m.responseView.Height = responseHeight - 2
}

func (m *Model) updateRequestEditor() {
	if m.selectedIdx < 0 || m.selectedIdx >= len(m.collection.Requests) {
		return
	}

	req := m.collection.Requests[m.selectedIdx]
	content := fmt.Sprintf("%s %s\n\nHeaders:\n", req.Method, req.URL)
	for k, v := range req.Headers {
		content += fmt.Sprintf("  %s: %s\n", k, v)
	}

	if req.Body != nil {
		bodyBytes, _ := collection.BodyToBytes(req.Body)
		formatted, _ := collection.FormatJSON(bodyBytes)
		content += fmt.Sprintf("\nBody:\n%s\n", formatted)
	}

	m.requestEditor.SetValue(content)
}

func (m Model) executeRequest() tea.Cmd {
	return func() tea.Msg {
		if m.selectedIdx < 0 || m.selectedIdx >= len(m.collection.Requests) {
			return nil
		}

		runner := collection.NewRunner(m.environment)
		runner.Resolver.LoadEnvironment(m.collection, m.environment)
		req := m.collection.Requests[m.selectedIdx]
		result := runner.ExecuteRequest(req)

		return executionResultMsg{result}
	}
}

type executionResultMsg struct {
	result collection.ExecutionResult
}

func (m *Model) handleExecutionResult(result collection.ExecutionResult) {
	m.results = append(m.results, result)

	formatted, _ := collection.FormatJSON(result.Response.Body)
	content := fmt.Sprintf("Status: %s\nTime: %v\nSize: %d bytes\n\n%s",
		result.Response.Status,
		result.Response.Time,
		result.Response.Size,
		formatted,
	)

	m.responseView.SetContent(content)
}

func (m Model) renderList() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62"))

	if m.activePane == paneList {
		style = style.BorderForeground(lipgloss.Color("170"))
	}

	return style.Render(m.requestList.View())
}

func (m Model) renderRequest() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62"))

	if m.activePane == paneRequest {
		style = style.BorderForeground(lipgloss.Color("170"))
	}

	return style.Render(m.requestEditor.View())
}

func (m Model) renderResponse() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62"))

	if m.activePane == paneResponse {
		style = style.BorderForeground(lipgloss.Color("170"))
	}

	return style.Render(m.responseView.View())
}

func (m Model) renderHelp() string {
	if m.showHelp {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Render("tab: next pane | shift+tab: prev pane | ctrl+e: execute | q: quit | ?: toggle help")
	}
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render("Press ? for help")
}
