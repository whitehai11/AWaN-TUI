package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/awan/awan-tui/api"
)

type panelMode int

const (
	modeChat panelMode = iota
	modeMemory
)

type runResponseMsg struct {
	response *api.AgentRunResponse
	err      error
}

type memoryResponseMsg struct {
	snapshot *api.MemorySnapshot
	err      error
}

type memoryEntry struct {
	Role    string
	Content string
}

type memoryView struct {
	ShortTerm []memoryEntry
	LongTerm  []memoryEntry
}

// Model is the Bubble Tea root model for AWaN-TUI.
type Model struct {
	client         *api.Client
	width          int
	height         int
	agents         []string
	selected       int
	input          textinput.Model
	messages       []ChatMessage
	memory         memoryView
	mode           panelMode
	status         string
	lastError      string
	busy           bool
	selectedModel  string
	leftStyle      lipgloss.Style
	centerStyle    lipgloss.Style
	bottomStyle    lipgloss.Style
	headerStyle    lipgloss.Style
	agentItemStyle lipgloss.Style
	activeItem     lipgloss.Style
	errorStyle     lipgloss.Style
}

// NewModel creates the root TUI model.
func NewModel(client *api.Client) Model {
	input := textinput.New()
	input.Placeholder = "Type a prompt and press Enter"
	input.Focus()
	input.CharLimit = 4000

	basePanel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1, 2)

	return Model{
		client:        client,
		agents:        []string{"default", "research-agent", "planner"},
		input:         input,
		status:        "Connecting to " + client.BaseURL(),
		selectedModel: "openai",
		leftStyle:     basePanel,
		centerStyle:   basePanel,
		bottomStyle:   basePanel,
		headerStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86")),
		agentItemStyle: lipgloss.NewStyle().
			PaddingLeft(1),
		activeItem: lipgloss.NewStyle().
			PaddingLeft(1).
			Foreground(lipgloss.Color("230")).
			Background(lipgloss.Color("29")).
			Bold(true),
		errorStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("203")).
			Bold(true),
	}
}

// Init starts the initial memory load.
func (m Model) Init() tea.Cmd {
	return m.fetchMemoryCmd()
}

// Update handles user input and async API messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up":
			if m.selected > 0 {
				m.selected--
			}
			m.status = "Selected agent: " + m.currentAgent()
			return m, m.fetchMemoryCmd()
		case "down":
			if m.selected < len(m.agents)-1 {
				m.selected++
			}
			m.status = "Selected agent: " + m.currentAgent()
			return m, m.fetchMemoryCmd()
		case "tab":
			if m.mode == modeChat {
				m.mode = modeMemory
				m.status = "Viewing memory for " + m.currentAgent()
			} else {
				m.mode = modeChat
				m.status = "Chat view for " + m.currentAgent()
			}
			return m, nil
		case "ctrl+r":
			m.status = "Refreshing memory for " + m.currentAgent()
			return m, m.fetchMemoryCmd()
		case "esc":
			m.input.SetValue("")
			m.lastError = ""
			return m, nil
		case "enter":
			if m.busy {
				return m, nil
			}

			prompt := strings.TrimSpace(m.input.Value())
			if prompt == "" {
				return m, nil
			}

			m.messages = append(m.messages, ChatMessage{
				Role:      "user",
				Content:   prompt,
				Timestamp: time.Now(),
			})
			m.input.SetValue("")
			m.busy = true
			m.lastError = ""
			m.mode = modeChat
			m.status = "Running prompt on " + m.currentAgent()

			return m, m.runAgentCmd(prompt)
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)

	switch msg := msg.(type) {
	case runResponseMsg:
		m.busy = false
		if msg.err != nil {
			m.lastError = msg.err.Error()
			m.status = "Runtime request failed"
			return m, nil
		}

		m.messages = append(m.messages, ChatMessage{
			Role:      "assistant",
			Content:   msg.response.Output,
			Timestamp: time.Now(),
		})
		m.status = fmt.Sprintf("Response received from %s", msg.response.Model)
		return m, m.fetchMemoryCmd()

	case memoryResponseMsg:
		if msg.err != nil {
			m.lastError = msg.err.Error()
			m.status = "Memory request failed"
			return m, nil
		}

		m.memory = memoryView{
			ShortTerm: mapMemory(msg.snapshot.ShortTerm),
			LongTerm:  mapMemory(msg.snapshot.LongTerm),
		}
		if !m.busy {
			m.status = "Connected to " + m.client.BaseURL()
		}
		return m, nil
	}

	return m, cmd
}

// View renders the full TUI.
func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading AWaN-TUI..."
	}

	leftWidth := max(22, m.width/4)
	centerWidth := max(40, m.width-leftWidth-4)
	bottomHeight := 5
	contentHeight := max(10, m.height-bottomHeight-2)

	left := m.renderAgents(leftWidth, contentHeight)
	center := m.renderCenter(centerWidth, contentHeight)
	bottom := m.renderBottom(m.width - 2)

	top := lipgloss.JoinHorizontal(lipgloss.Top, left, center)
	return lipgloss.JoinVertical(lipgloss.Left, top, bottom)
}

func (m Model) renderAgents(width, height int) string {
	var items []string
	for i, agent := range m.agents {
		label := agent
		if i == m.selected {
			label = m.activeItem.Render("> " + label)
		} else {
			label = m.agentItemStyle.Render("  " + label)
		}
		items = append(items, label)
	}

	body := strings.Join(items, "\n")
	content := m.headerStyle.Render("Agents") + "\n\n" + body
	return m.leftStyle.Width(width).Height(height).Render(content)
}

func (m Model) renderCenter(width, height int) string {
	title := "Chat"
	body := formatChat(m.messages, width-8)
	if m.mode == modeMemory {
		title = "Memory"
		body = formatMemory(m.memory, width-8)
	}

	content := m.headerStyle.Render(title) + "\n\n" + body
	return m.centerStyle.Width(width).Height(height).Render(content)
}

func (m Model) renderBottom(width int) string {
	help := "Enter: send  Tab: toggle chat/memory  Ctrl+R: refresh memory  Up/Down: select agent  Q: quit"
	status := m.status
	if m.lastError != "" {
		status = m.errorStyle.Render(m.lastError)
	}

	content := m.headerStyle.Render("Command Input") +
		"\n\n" + m.input.View() +
		"\n\n" + status +
		"\n" + help

	return m.bottomStyle.Width(width).Height(5).Render(content)
}

func (m Model) runAgentCmd(prompt string) tea.Cmd {
	agent := m.currentAgent()
	model := m.selectedModel

	return func() tea.Msg {
		response, err := m.client.RunAgent(api.AgentRunRequest{
			Agent:  agent,
			Model:  model,
			Prompt: prompt,
		})

		return runResponseMsg{
			response: response,
			err:      err,
		}
	}
}

func (m Model) fetchMemoryCmd() tea.Cmd {
	agent := m.currentAgent()
	return func() tea.Msg {
		snapshot, err := m.client.GetMemory(agent)
		return memoryResponseMsg{
			snapshot: snapshot,
			err:      err,
		}
	}
}

func (m Model) currentAgent() string {
	if m.selected < 0 || m.selected >= len(m.agents) {
		return "default"
	}
	return m.agents[m.selected]
}

func mapMemory(records []api.MemoryRecord) []memoryEntry {
	entries := make([]memoryEntry, 0, len(records))
	for _, record := range records {
		entries = append(entries, memoryEntry{
			Role:    record.Role,
			Content: record.Content,
		})
	}
	return entries
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
