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

type GemtextRenderer struct {
	tabMode TabMode
	sb      *strings.Builder
	hints   map[*gemtext.Node]string
}

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

func renderGemtext(node *gemtext.Node, hints map[*gemtext.Node]string, tabMode TabMode) string {
	var sb strings.Builder
	r := &GemtextRenderer{tabMode: tabMode, sb: &sb, hints: hints}
	r.render(node)
	return r.sb.String()
}

func (r *GemtextRenderer) render(node *gemtext.Node) {
	var content string

	switch node.Type {
	case gemtext.NodeText:
		content = textStyle.Render(wrapText(node.RawContent, 0))

	case gemtext.NodeLink:
		content = r.renderLink(node)

	case gemtext.NodeBlockquote:
		content = blockQuotesStyle.Render(wrapText(node.RawContent, 2))

	case gemtext.NodeHeading1, gemtext.NodeHeading2, gemtext.NodeHeading3:
		content = headingStyle.Render(wrapText(node.RawContent, 0))
	}

	r.sb.WriteString("\n")
	r.sb.WriteString(content)

	for _, child := range node.Children {
		r.render(child)
	}
}

func (r *GemtextRenderer) renderLink(node *gemtext.Node) string {
	meta := node.Meta.(gemtext.LinkMeta)
	style := linkStyle
	prefix := "-> "
	if r.tabMode == SelectLink {
		if hint, ok := r.hints[node]; ok {
			style = linkJumperStyle
			prefix = hint
		}
	}
	return fmt.Sprintf("%s%s - %s",
		style.Underline(true).Render(prefix),
		style.Render(" "+meta.Url),
		style.Render(meta.Label))
}

func wrapText(text string, leftPadding int) string {
	available := max(20, 100-leftPadding)

	wrapped := wordwrap.String(text, available)
	wrapped = wrap.String(wrapped, available)
	return wrapped
}
