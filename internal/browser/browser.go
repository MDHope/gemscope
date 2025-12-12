package browser

import (
	"fmt"
	"os"
	"strings"

	gemini_client "github.com/MDHope/gemscope/internal/gemini-client"
	"github.com/MDHope/gemscope/internal/gemtext"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	viewportHeightOffset = 15
	viewportWidthPadding = 2
	viewportYPosition    = 5
	scrollStep           = 5
	maxUrlLength         = 156

	bookmaksTitle = "Bookmarks"
)

type model struct {
	activeTab    int
	tabs         []*tab
	width        int
	height       int
	geminiClient *gemini_client.GeminiClient
	hintMode     *hintMode
	bookmarks    *Bookmarks
	keys         keyMap
	showFullHelp bool
}

func Start(geminiClient *gemini_client.GeminiClient) {
	p := tea.NewProgram(initialModel(geminiClient), tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Printf("UI init error: %v\n", err)
		os.Exit(1)
	}
}

func initialModel(geminiClient *gemini_client.GeminiClient) model {
	b, _ := LoadBookmarks()
	m := model{
		activeTab:    0,
		tabs:         []*tab{},
		geminiClient: geminiClient,
		bookmarks:    b,
		keys:         keys,
	}
	m = m.appendNewTab()
	return m
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	at := m.getActiveTab()
	currentMode := at.mode

	switch msg := msg.(type) {
	case editorFinishedMsg:
		if msg.err != nil {
			m.updateTabContent(fmt.Sprintf("Failed to edit: %s\n", msg.err.Error()))
			return m, nil
		}

		m.bookmarks.loadBookmarksFromFile()
		m.setPage(&gemini_client.GeminiResponse{Body: m.bookmarks.BookmarksContent}, bookmaksTitle)
		return m, nil
	case navigateMsg:
		return m.handleNavigation(msg.Url)
	case tea.KeyMsg:
		if msg.String() == "ctrl+q" {
			return m, tea.Quit
		}

		switch currentMode {
		case SelectLink:
			return m.handleSelectLinkMode(msg)
		case Insert, New:
			return m.handleInsertMode(msg)
		case View:
			return m.handleViewMode(msg)
		case ViewBookmarks:
			return m.handleViewBookmarksMode(msg)
		}
	case tea.WindowSizeMsg:
		return m.handleResize(msg)
	}

	at.viewport, cmd = at.viewport.Update(msg)
	m.tabs[m.activeTab] = at

	return m, cmd
}

func (m model) View() string {
	doc := strings.Builder{}
	at := m.getActiveTab()
	isUrlFocused := at.urlInput.Focused()

	renderedTabBar := m.renderTabBar()
	doc.WriteString(renderedTabBar)

	doc.WriteString(at.contentUrlInputView(isUrlFocused))

	doc.WriteString(fmt.Sprintf("%s\n%s\n%s", at.contentHeaderView(), at.viewport.View(), at.contentFooterView()))

	doc.WriteString(m.renderHelp())

	return docStyle.Render(doc.String())
}

func (m model) handleResize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	updatedTabs := []*tab{}

	for _, tab := range m.tabs {
		tab.viewport.Width = msg.Width - viewportWidthPadding
		tab.viewport.Height = msg.Height - viewportHeightOffset
		tab.viewport.YPosition = viewportYPosition
		tab.urlInput.Width = msg.Width
		updatedTabs = append(updatedTabs, tab)
	}

	m.width = msg.Width
	m.height = msg.Height
	m.tabs = updatedTabs
	return m, nil
}

func (m model) handleViewBookmarksMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Bookmarks.QuitMode):
		m.changeActiveTabMode(View)
		history := m.getActiveTab().history
		if len(history) > 0 {
			item := history[len(history)-1]
			m.setPage(item.response, item.url)
		} else {
			m.setPage(&gemini_client.GeminiResponse{Body: "\n"}, "New tab")
		}
		return m, nil
	case key.Matches(msg, m.keys.Bookmarks.Editor):
		return m, m.bookmarks.openEditor()
	case key.Matches(msg, m.keys.Global.Help):
		m.showFullHelp = !m.showFullHelp
		return m, nil
	case key.Matches(msg, m.keys.Bookmarks.HintLinks):
		m.changeActiveTabMode(SelectLink)
		m.updateTabContent(renderGemtext(m.tabs[m.activeTab].parsed, m.tabs[m.activeTab].hints, SelectLink))
		m.hintMode = &hintMode{}
		return m, nil
	}

	return m, nil
}

func (m model) handleViewMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.View.CloseTab):
		m = m.closeActiveTab()
		return m, nil
	case key.Matches(msg, m.keys.View.Bookmarks):
		m.changeActiveTabMode(ViewBookmarks)
		m.setPage(&gemini_client.GeminiResponse{Body: m.bookmarks.BookmarksContent}, bookmaksTitle)
		return m, nil
	case key.Matches(msg, m.keys.Global.Help):
		m.showFullHelp = !m.showFullHelp
		return m, nil
	case key.Matches(msg, m.keys.View.GoBack):
		m.goBack()
		return m, nil
	case key.Matches(msg, m.keys.View.FocusUrl):
		m.changeActiveTabMode(Insert)
		cmd := m.tabs[m.activeTab].urlInput.Focus()
		return m, cmd
	case key.Matches(msg, m.keys.View.ScrollDown):
		m.tabs[m.activeTab].viewport.ScrollDown(scrollStep)
		return m, nil
	case key.Matches(msg, m.keys.View.ScrollUp):
		m.tabs[m.activeTab].viewport.ScrollUp(scrollStep)
		return m, nil
	case key.Matches(msg, m.keys.View.PrevTab):
		m = m.previousTab()
		return m, nil
	case key.Matches(msg, m.keys.View.NextTab):
		m = m.nextTab()
		return m, nil
	case key.Matches(msg, m.keys.View.NewTab):
		m = m.appendNewTab()
		return m, nil
	case key.Matches(msg, m.keys.View.HintLinks):
		m.changeActiveTabMode(SelectLink)
		m.updateTabContent(renderGemtext(m.tabs[m.activeTab].parsed, m.tabs[m.activeTab].hints, SelectLink))
		m.hintMode = &hintMode{}
		return m, nil
	}

	return m, nil
}

func (m model) handleInsertMode(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(msg, m.keys.Insert.Submit):
			inputValue := m.tabs[m.activeTab].urlInput.Value()
			m.loadPage(inputValue)
			return m, nil
		case key.Matches(msg, m.keys.Insert.QuitMode):
			m.changeActiveTabMode(View)
			m.tabs[m.activeTab].urlInput.Blur()
			return m, nil
		}
	}

	var cmd tea.Cmd
	at := m.getActiveTab()
	at.urlInput, cmd = at.urlInput.Update(msg)
	return m, cmd
}

func (m model) handleSelectLinkMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	char := msg.String()

	switch {
	case key.Matches(msg, m.keys.Links.QuitMode):
		if m.tabs[m.activeTab].title == "Bookmarks" {
			m.changeActiveTabMode(ViewBookmarks)
		} else {
			m.changeActiveTabMode(View)
		}
		m.updateTabContent(renderGemtext(m.tabs[m.activeTab].parsed, m.tabs[m.activeTab].hints, View))
		return m, nil
	case key.Matches(msg, m.keys.Global.Help):
		m.showFullHelp = !m.showFullHelp
		return m, nil
	case char == "backspace":
		if len(m.hintMode.input) > 0 {
			m.hintMode.input = m.hintMode.input[:len(m.hintMode.input)-1]
		}
	default:
		key := msg.String()
		if len(char) == 1 && strings.Contains(hintChars, key) {
			m.hintMode.input += char
		}
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

func (m model) loadPage(url string) {
	res, err := m.geminiClient.Fetch(url)

	newHistoryItem := &historyItem{}
	newHistoryItem.response = res
	newHistoryItem.url = url
	m.appendToHistory(newHistoryItem)

	m.setUrlInput(url)

	if err != nil {
		m.updateTabContent(err.Error())
		return
	}

	m.setPage(res, url)
	m.changeActiveTabMode(View)
}

func (m model) setPage(res *gemini_client.GeminiResponse, url string) {
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

	m.tabs[m.activeTab].parsed = parsedRes
	m.tabs[m.activeTab].links = collectedLinks
	m.tabs[m.activeTab].hints = hints

	m.updateTabContent(newTabContent)
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
