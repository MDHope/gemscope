package ui

import (
	"fmt"
	"os"
	"strings"

	gemini_client "github.com/MDHope/gemscope/internal/gemini-client"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	activeTab    int
	tabs         []tab
	width        int
	height       int
	geminiClient *gemini_client.GeminiClient
}

func InitUI(geminiClient *gemini_client.GeminiClient) {
	p := tea.NewProgram(initialModel(geminiClient), tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Printf("UI init error: %v\n", err)
		os.Exit(1)
	}
}

func initialModel(geminiClient *gemini_client.GeminiClient) model {

	m := model{
		activeTab:    0,
		tabs:         []tab{},
		geminiClient: geminiClient,
	}
	m = m.appendNewTab()
	return m
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	currentMode := m.getActiveTab().mode

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+q":
			return m, tea.Quit
		case "ctrl+c":
			m = m.closeActiveTab()
			return m, nil
		case "esc":
			m = m.changeActiveTabMode(View)
			return m, nil
		case "L":
			m = m.changeActiveTabMode(Insert)
			cmd := m.tabs[m.activeTab].urlInput.Focus()
			return m, cmd
		case "j":
			if currentMode == View {
				m.tabs[m.activeTab].viewport.ScrollDown(5)
				return m, nil
			}
		case "k":
			if currentMode == View {
				m.tabs[m.activeTab].viewport.ScrollUp(5)
				return m, nil
			}
		case "{":
			if currentMode == View {
				m = m.previousTab()
				return m, nil
			}
		case "}":
			if currentMode == View {
				m = m.nextTab()
				return m, nil
			}
		case "ctrl+t":
			m = m.appendNewTab()
			return m, nil
		case "enter":
			inputValue := m.tabs[m.activeTab].urlInput.Value()
			res, err := m.geminiClient.Fetch(inputValue)
			if err != nil {
				return m, nil
			}

			var newTabContent string

			if len(res.Body) > 0 {
				newTabContent = contentResponseView(res.Body, m.tabs[m.activeTab].mode)
			} else {
				newTabContent = fmt.Sprintf("%d: %s\n", res.Status, res.Meta)
			}

			m = m.updateTabContent(newTabContent)
			m = m.changeActiveTabMode(View)
			m.tabs[m.activeTab].title = getTabTitle(inputValue)
			m.tabs[m.activeTab].urlInput.Blur()
			m.tabs[m.activeTab].viewport.ScrollUp(100)
			return m, nil

		case "f":
			m = m.changeActiveTabMode(SelectLink)
			return m, nil
		}
	case tea.WindowSizeMsg:
		updatedTabs := []tab{}

		for _, tab := range m.tabs {
			tab.viewport.Width = msg.Width - 2
			tab.viewport.Height = msg.Height - 15
			tab.viewport.YPosition = 5
			tab.viewport.SetContent(tab.content)
			tab.urlInput.Width = msg.Width
			updatedTabs = append(updatedTabs, tab)
		}

		m.width = msg.Width
		m.height = msg.Height
		m.tabs = updatedTabs
		return m, nil
	}

	var cmdInput tea.Cmd
	var cmdViewport tea.Cmd

	at := m.getActiveTab()

	if currentMode == Insert || currentMode == New {
		at.urlInput, cmdInput = at.urlInput.Update(msg)
		cmds = append(cmds, cmdInput)
	}

	at.viewport, cmdViewport = at.viewport.Update(msg)
	m.tabs[m.activeTab] = at

	cmds = append(cmds, cmdViewport)
	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	doc := strings.Builder{}
	at := m.getActiveTab()

	renderedTabBar := m.renderTabBar()
	doc.WriteString(renderedTabBar)

	doc.WriteString(at.contentUrlInputView())

	doc.WriteString(fmt.Sprintf("%s\n%s\n%s", at.contentHeaderView(), at.viewport.View(), at.contentFooterView()))

	return docStyle.Render(doc.String())
	// line := lipgloss.JoinHorizontal(lipgloss.Bottom, strings.Repeat("-", 50))
	// doc.WriteString(row)
	// doc.WriteString(line)
	// doc.WriteString("\n")
	// doc.WriteString(windowStyle.Width((lipgloss.Width(row) - windowStyle.GetHorizontalFrameSize())).Render(getTabTitle(m.activeTab)))
}
