package browser

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

var (
	helpKeyStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	helpDescStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	helpSepStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
)

type keyMap struct {
	Global    GlobalKeys
	New       ViewModeKeys
	View      ViewModeKeys
	Insert    InsertModeKeys
	Bookmarks BookmarksModeKeys
	Links     LinksModeKeys
}

type GlobalKeys struct {
	Quit key.Binding
	Help key.Binding
}

type ViewModeKeys struct {
	NewTab     key.Binding
	CloseTab   key.Binding
	PrevTab    key.Binding
	NextTab    key.Binding
	Bookmarks  key.Binding
	GoBack     key.Binding
	FocusUrl   key.Binding
	ScrollUp   key.Binding
	ScrollDown key.Binding
	HintLinks  key.Binding
}

type InsertModeKeys struct {
	QuitMode key.Binding
	Submit   key.Binding
}

type BookmarksModeKeys struct {
	Editor    key.Binding
	HintLinks key.Binding
	QuitMode  key.Binding
}

type LinksModeKeys struct {
	QuitMode key.Binding
}

var hintLinks = key.NewBinding(
	key.WithKeys("f"),
	key.WithHelp("f", "hint links"),
)
var quitMode = key.NewBinding(
	key.WithKeys("esc"),
	key.WithHelp("esc", "quit mode"),
)

var viewModeKeys = ViewModeKeys{
	CloseTab: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "close tab"),
	),
	NewTab: key.NewBinding(
		key.WithKeys("ctrl+t"),
		key.WithHelp("ctrl+t", "new tab"),
	),
	PrevTab: key.NewBinding(
		key.WithKeys("{"),
		key.WithHelp("{", "prev tab"),
	),
	NextTab: key.NewBinding(
		key.WithKeys("}"),
		key.WithHelp("}", "next tab"),
	),
	Bookmarks: key.NewBinding(
		key.WithKeys("ctrl+b"),
		key.WithHelp("ctrl+b", "bookmarks"),
	),
	GoBack: key.NewBinding(
		key.WithKeys("H"),
		key.WithHelp("H", "go back"),
	),
	FocusUrl: key.NewBinding(
		key.WithKeys("L"),
		key.WithHelp("L", "focus url"),
	),
	ScrollUp: key.NewBinding(
		key.WithKeys("k"),
		key.WithHelp("k", "scroll up"),
	),
	ScrollDown: key.NewBinding(
		key.WithKeys("j"),
		key.WithHelp("j", "scroll down"),
	),
	HintLinks: key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("f", "hint links"),
	),
}

var globalKeys = GlobalKeys{
	Quit: key.NewBinding(
		key.WithKeys("cltr+q"),
		key.WithHelp("ctrl+q", "quit"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
}

var keys = keyMap{
	Global: globalKeys,
	New:    viewModeKeys,
	View:   viewModeKeys,
	Insert: InsertModeKeys{
		QuitMode: quitMode,
		Submit: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "submit"),
		),
	},
	Bookmarks: BookmarksModeKeys{
		Editor: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "editor"),
		),
		HintLinks: hintLinks,
		QuitMode:  quitMode,
	},
	Links: LinksModeKeys{
		QuitMode: quitMode,
	},
}

func (m model) renderHelp() string {
	mode := m.getActiveTab().mode
	keys := m.keys

	var all []key.Binding
	all = append(all, keys.Global.Quit, keys.Global.Help)

	if !m.showFullHelp {
		return fmt.Sprintf("\n%s", renderBinding(all))
	}

	var specific []key.Binding

	switch mode {
	case New, View:
		specific = append(specific, keys.View.NewTab, keys.View.CloseTab, keys.View.PrevTab, keys.View.NextTab, keys.View.Bookmarks, keys.View.GoBack, keys.View.FocusUrl, keys.View.ScrollUp, keys.View.ScrollDown, keys.View.HintLinks)
		break
	case Insert:
		specific = append(specific, keys.Insert.Submit, keys.Insert.QuitMode)
		break
	case SelectLink:
		specific = append(specific, keys.Links.QuitMode)
		break
	case ViewBookmarks:
		specific = append(specific, keys.Bookmarks.Editor, keys.Bookmarks.HintLinks, keys.Bookmarks.QuitMode)
		break
	}

	return fmt.Sprintf("\n%s%s", renderBinding(all), renderBinding(specific))
}

func renderBinding(bindings []key.Binding) string {
	var parts []string

	for _, b := range bindings {
		if !b.Enabled() {
			continue
		}

		help := b.Help()
		part := helpKeyStyle.Render(help.Key) + " " + helpDescStyle.Render(help.Desc)
		parts = append(parts, part)
	}

	sep := helpSepStyle.Render(" • ")
	return fmt.Sprintf("\n%s", strings.Join(parts, sep))
}
