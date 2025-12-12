package browser

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
)

const bookmarksDefaultContent = "# Bookmarks\n\n"

type editorFinishedMsg struct {
	err error
}

type Bookmarks struct {
	BookmarksContent string
	filePath         string
}

func LoadBookmarks() (*Bookmarks, error) {
	b := &Bookmarks{}

	homeDir, _ := os.UserHomeDir()
	bookmarksFile := filepath.Join(homeDir, "/gemscope/", "gemscope_bookmarks.gmi")
	dir := filepath.Dir(bookmarksFile)
	if err := os.Mkdir(dir, 0755); err != nil && !os.IsExist(err) {
		return nil, fmt.Errorf("failed to create dir: %w", err)
	}

	b.filePath = bookmarksFile

	err := b.loadBookmarksFromFile()
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (b *Bookmarks) loadBookmarksFromFile() error {
	file, err := os.ReadFile(b.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(filepath.Dir(b.filePath), 0755); err != nil {
				return err
			}
			if err := os.WriteFile(b.filePath, []byte(bookmarksDefaultContent), 0644); err != nil {
				return err
			}

			b.BookmarksContent = bookmarksDefaultContent
			return nil
		}

		return fmt.Errorf("err: %w", err)
	}

	b.BookmarksContent = string(file)
	return nil
}

func (b *Bookmarks) openEditor() tea.Cmd {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "nvim"
	}

	cmd := exec.Command(editor, b.filePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		return editorFinishedMsg{err: err}
	})
}
