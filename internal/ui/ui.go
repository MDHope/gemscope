package ui

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	gemini_client "github.com/MDHope/gemscope/internal/gemini-client"
	"github.com/MDHope/gemscope/internal/gemtext"
	tea "github.com/charmbracelet/bubbletea"
)

type navigateMsg struct {
	Url string
}

type hintMatch struct {
	hint string
	node *gemtext.Node
}

type hintMode struct {
	input   string
	matches []hintMatch
}

type model struct {
	activeTab    int
	tabs         []*tab
	width        int
	height       int
	geminiClient *gemini_client.GeminiClient
	hintMode     *hintMode
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
		tabs:         []*tab{},
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
	case navigateMsg:
		return m.handleNavigation(msg.Url)
	case tea.KeyMsg:
		if currentMode == SelectLink {
			return m.handleHintInput(msg)
		}

		switch msg.String() {
		case "ctrl+q":
			return m, tea.Quit
		case "ctrl+c":
			m = m.closeActiveTab()
			return m, nil
		case "esc":
			if currentMode == SelectLink {
				m.updateTabContent(renderGemtext(m.tabs[m.activeTab].parsed, m.tabs[m.activeTab].hints, View))
			}
			m.changeActiveTabMode(View)
			return m, nil
		case "L":
			m.changeActiveTabMode(Insert)
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
			m.loadPage(inputValue)
			return m, nil

		case "f":
			if currentMode == View {
				m.changeActiveTabMode(SelectLink)
				m.updateTabContent(renderGemtext(m.tabs[m.activeTab].parsed, m.tabs[m.activeTab].hints, SelectLink))
				m.hintMode = &hintMode{}
				return m, nil
			}
		}
	case tea.WindowSizeMsg:
		updatedTabs := []*tab{}

		for _, tab := range m.tabs {
			tab.viewport.Width = msg.Width - 2
			tab.viewport.Height = msg.Height - 15
			tab.viewport.YPosition = 5
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

func (m model) handleHintInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	if key == "esc" {
		m.changeActiveTabMode(View)
		m.updateTabContent(renderGemtext(m.tabs[m.activeTab].parsed, m.tabs[m.activeTab].hints, View))
		return m, nil
	}

	if key == "backspace" && len(m.hintMode.input) > 0 {
		m.hintMode.input = m.hintMode.input[:len(m.hintMode.input)-1]
	} else if len(key) == 1 && strings.Contains(hintChars, key) {
		m.hintMode.input += key
	}

	m.hintMode.matches = nil

	for node, hint := range m.tabs[m.activeTab].hints {
		if strings.HasPrefix(hint, m.hintMode.input) {
			m.hintMode.matches = append(m.hintMode.matches, hintMatch{hint: hint, node: node})
		}
	}

	if len(m.hintMode.matches) == 1 && m.hintMode.matches[0].hint == m.hintMode.input {
		linkNode := m.hintMode.matches[0].node
		m.changeActiveTabMode(View)
		m.updateTabContent(renderGemtext(m.tabs[m.activeTab].parsed, m.tabs[m.activeTab].hints, View))
		return m, m.navigateTo(linkNode)
	}

	return m, nil
}

func (m model) navigateTo(node *gemtext.Node) tea.Cmd {
	linkMeta, ok := node.Meta.(gemtext.LinkMeta)
	if !ok {
		return nil
	}

	return func() tea.Msg {
		return navigateMsg{Url: linkMeta.Url}
	}
}

func (m model) handleNavigation(href string) (tea.Model, tea.Cmd) {
	if !strings.HasPrefix(href, "gemini://") {
		at := m.getActiveTab()
		currHref := at.url
		base, err := url.Parse(currHref)
		if err != nil {
			return m, nil
		}

		ref, err := url.Parse(href)
		if err != nil {
			return m, nil
		}

		href = base.ResolveReference(ref).String()
	}

	m.loadPage(href)
	return m, nil
}

func (m model) View() string {
	doc := strings.Builder{}
	at := m.getActiveTab()

	renderedTabBar := m.renderTabBar()
	doc.WriteString(renderedTabBar)

	doc.WriteString(at.contentUrlInputView())

	doc.WriteString(fmt.Sprintf("%s\n%s\n%s", at.contentHeaderView(), at.viewport.View(), at.contentFooterView()))

	return docStyle.Render(doc.String())
}

func (m model) loadPage(url string) {
	res, err := m.geminiClient.Fetch(url)
	if err != nil {
		return
	}

	var newTabContent string
	parsedRes := gemtext.Parse(res.Body)

	collectedLinks := collectLinks(parsedRes)
	hintKeys := generateHints(len(collectedLinks))

	hints := make(map[*gemtext.Node]string, len(collectedLinks))
	for i, node := range collectedLinks {
		hints[node] = hintKeys[i]
	}

	if len(res.Body) > 0 {
		newTabContent = renderGemtext(parsedRes, hints, m.tabs[m.activeTab].mode)
	} else {
		newTabContent = fmt.Sprintf("%d: %s\n", res.Status, res.Meta)
	}

	m.updateTabContent(newTabContent)
	m.tabs[m.activeTab].parsed = parsedRes
	m.tabs[m.activeTab].url = url
	m.tabs[m.activeTab].links = collectedLinks
	m.tabs[m.activeTab].hints = hints
	m.tabs[m.activeTab].urlInput.SetValue(url)
	m.tabs[m.activeTab].urlInput.SetCursor(len(url))
	m.changeActiveTabMode(View)
	m.tabs[m.activeTab].title = getTabTitle(url)
	m.tabs[m.activeTab].urlInput.Blur()
	m.tabs[m.activeTab].viewport.ScrollUp(100)
}

func collectLinks(node *gemtext.Node) []*gemtext.Node {
	var links []*gemtext.Node

	collectLinksRecursive(node, &links)

	return links
}

func collectLinksRecursive(node *gemtext.Node, links *[]*gemtext.Node) {
	if meta, ok := node.Meta.(gemtext.LinkMeta); ok {
		if strings.HasPrefix(meta.Url, "gemini://") || !strings.Contains(meta.Url, "//") {
			*links = append(*links, node)
		}
	}

	for _, child := range node.Children {
		collectLinksRecursive(child, links)
	}
}
