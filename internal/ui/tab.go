package ui

import (
	"slices"
	"strings"

	"github.com/MDHope/gemscope/internal/gemtext"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

var (
	highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}

	activeTabBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      " ",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┘",
		BottomRight: "└",
	}

	tabBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┴",
		BottomRight: "┴",
	}

	tabStyle = lipgloss.NewStyle().
			Border(tabBorder, true).
			BorderForeground(highlight).
			Padding(0, 1)

	activeTabStyle = tabStyle.Border(activeTabBorder, true)

	tabGap = tabStyle.
		BorderTop(false).
		BorderLeft(false).
		BorderRight(false)

	docStyle = lipgloss.NewStyle().Padding(1, 2, 1, 2)
)

type TabMode int

const (
	New        TabMode = 0
	Insert     TabMode = 1
	View       TabMode = 2
	SelectLink TabMode = 3
)

type tab struct {
	mode     TabMode
	url      string
	urlInput textinput.Model
	viewport viewport.Model
	title    string
	content  string
	parsed   *gemtext.Node
	links    []*gemtext.Node
	hints    map[*gemtext.Node]string
}

func (m model) newTab() tab {
	ti := textinput.New()
	ti.Placeholder = "Enter url..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = m.width

	vp := viewport.New(m.width, m.height-15)
	vp.YPosition = 20
	vp.SetContent("")

	return tab{
		mode:     New,
		urlInput: ti,
		content:  "",
		parsed:   nil,
		links:    nil,
		hints:    nil,
		viewport: vp,
		title:    "New tab",
	}
}

func (m model) renderTabBar() string {
	var renderedTabs []string

	for idx, tab := range m.tabs {
		var style lipgloss.Style
		isActive := idx == m.activeTab

		if isActive {
			style = activeTabStyle
		} else {
			style = tabStyle
		}

		renderedTabs = append(renderedTabs, style.Render(tab.title))
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	gap := tabGap.Render(strings.Repeat(" ", max(0, m.width-lipgloss.Width(row)-2)))
	row = lipgloss.JoinHorizontal(lipgloss.Bottom, row, gap)

	return row + "\n\n"
}

func (m model) getActiveTab() *tab {
	return m.tabs[m.activeTab]
}

func getTabTitle(url string) string {
	url = strings.TrimPrefix(url, "gemini://")
	url = strings.Replace(url, "/", "", -1)
	if len(url) > 10 {
		parts := strings.Split(url, "")
		url = strings.Join(parts[0:10], "") + "..."
		return url
	} else {
		return url
	}
}

func (m model) getPreviousTabId() int {
	return max(m.activeTab-1, 0)
}

func (m model) previousTab() model {
	id := max(m.activeTab-1, 0)
	m.activeTab = id
	return m
}

func (m model) nextTab() model {
	id := min(m.activeTab+1, len(m.tabs)-1)
	m.activeTab = id
	return m
}

func (m model) appendNewTab() model {
	t := m.newTab()

	m.tabs = append(m.tabs, &t)
	m.activeTab = len(m.tabs) - 1
	return m
}

func (m model) closeActiveTab() model {
	if len(m.tabs) < 2 {
		return m
	}

	m.tabs = slices.Delete(m.tabs, m.activeTab, m.activeTab+1)
	m.activeTab = m.getPreviousTabId()
	return m
}

func (m model) changeActiveTabMode(mode TabMode) {
	m.tabs[m.activeTab].mode = mode
}

func (m model) updateTabContent(content string) {
	m.tabs[m.activeTab].content = content
	m.tabs[m.activeTab].viewport.SetContent(content)
}
