package ui

import (
	"fmt"
	"strings"

	"github.com/MDHope/gemscope/internal/gemtext"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
	"github.com/muesli/reflow/wrap"
)

var (
	// colors
	special = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}

	contentTitleStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1)
	}()

	contentInfoStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return contentTitleStyle.BorderStyle(b)
	}()

	textStyle = func() lipgloss.Style {
		return lipgloss.NewStyle()
	}()
	headingStyle = func() lipgloss.Style {
		return lipgloss.NewStyle().Bold(true)
	}()
	blockQuotesStyle = func() lipgloss.Style {
		return lipgloss.NewStyle()
	}()
	linkJumperStyle = func() lipgloss.Style {
		return lipgloss.NewStyle().Foreground(special)
	}()
	linkStyle = func() lipgloss.Style {
		return lipgloss.NewStyle()
	}()
	linkJumpStyle = func() lipgloss.Style {
		return lipgloss.NewStyle()
	}()
)

func (t tab) contentUrlInputView() string {
	return fmt.Sprintf(t.urlInput.View() + "\n\n")
}

func (t tab) contentHeaderView() string {
	// title := contentTitleStyle.Render("Mr. Pager")
	line := strings.Repeat("─", max(0, t.viewport.Width))
	return lipgloss.JoinHorizontal(lipgloss.Center, line)
}

func (t tab) contentFooterView() string {
	info := contentInfoStyle.Render(fmt.Sprintf("%3.f%%", t.viewport.ScrollPercent()*100))
	line := strings.Repeat("─", max(0, t.viewport.Width-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}

func contentResponseView(raw string, tabMode TabMode) string {
	parsed := gemtext.Parse(raw)

	return renderContent(parsed, tabMode)
}

func renderContent(node *gemtext.Node, tabMode TabMode) string {
	var sb strings.Builder
	render(node, &sb, tabMode)
	return sb.String()
}

func render(node *gemtext.Node, sb *strings.Builder, tabMode TabMode) {
	var content string

	switch node.Type {
	case gemtext.NodeText:
		content = textStyle.Render(wrapText(node.RawContent, 0))

	case gemtext.NodeLink:
		if tabMode == SelectLink {
			content = fmt.Sprintf("%s - %s", linkJumperStyle.Underline(true).Render("aa "+node.Meta.(gemtext.LinkMeta).Url), linkJumperStyle.Render(node.Meta.(gemtext.LinkMeta).Label))
		} else {
			content = fmt.Sprintf("%s - %s", linkStyle.Underline(true).Render("-> "+node.Meta.(gemtext.LinkMeta).Url), linkStyle.Render(node.Meta.(gemtext.LinkMeta).Label))
		}

	case gemtext.NodeBlockquote:
		content = blockQuotesStyle.Render(wrapText(node.RawContent, 2))
	case gemtext.NodeHeading1, gemtext.NodeHeading2, gemtext.NodeHeading3:
		content = headingStyle.Render(wrapText(node.RawContent, 0))
	}

	sb.WriteString("\n")

	sb.WriteString(content)

	for _, child := range node.Children {
		render(child, sb, tabMode)
	}
}

func wrapText(text string, leftPadding int) string {
	available := max(20, 100-leftPadding)

	wrapped := wordwrap.String(text, available)
	wrapped = wrap.String(wrapped, available)
	return wrapped
}
