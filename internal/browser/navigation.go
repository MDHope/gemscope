package browser

import (
	"net/url"
	"strings"

	"github.com/MDHope/gemscope/internal/gemtext"
	tea "github.com/charmbracelet/bubbletea"
)

type navigateMsg struct {
	Url string
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
		h := m.getActiveHistoryItem()
		currHref := h.url
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
