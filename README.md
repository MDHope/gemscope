# gemscope

A terminal-based [Gemini protocol](https://geminiprotocol.net/) browser built with [Bubble Tea](https://github.com/charmbracelet/bubbletea).

## Features

- Multi-tab browsing with tab bar
- Keyboard-driven navigation with hint-based link selection (vimium-style)
- Gemtext rendering (headings, links, preformatted blocks, blockquotes, lists)
- Bookmarks stored as a `.gmi` file, editable in `$EDITOR`
- Per-tab navigation history (up to 5 entries)
- TLS certificate pinning via `~/.gemini_known_hosts`

## Installation

```sh
task build        # produces ./gemscope binary
```

Or manually:

```sh
go build -o gemscope cmd/main.go
```

## Usage

```sh
./gemscope
```

On first run, gemscope creates `~/gemscope/` and stores:
- `.gemini_known_hosts` — TLS certificate fingerprints
- `gemscope_bookmarks.gmi` — bookmarks file

## Keybindings

### Global
| Key | Action |
|-----|--------|
| `ctrl+q` | Quit |
| `?` | Toggle help |

### View mode
| Key | Action |
|-----|--------|
| `L` | Focus URL bar |
| `H` | Go back |
| `j` / `k` | Scroll down / up |
| `f` | Activate link hints |
| `ctrl+t` | New tab |
| `ctrl+c` | Close tab |
| `{` / `}` | Previous / next tab |
| `ctrl+b` | Open bookmarks |

### URL bar (Insert mode)
| Key | Action |
|-----|--------|
| `Enter` | Navigate to URL |
| `Esc` | Cancel |

### Link hint mode
Type the highlighted letter combination to follow a link. `Esc` to cancel.

### Bookmarks
| Key | Action |
|-----|--------|
| `e` | Open bookmarks file in `$EDITOR` |
| `f` | Activate link hints |
| `Esc` | Back |

## Data files

| Path | Purpose |
|------|---------|
| `~/gemscope/.gemini_known_hosts` | Certificate fingerprint store |
| `~/gemscope/gemscope_bookmarks.gmi` | Bookmarks (standard gemtext format) |

## Known limitations / TODO

- Status 10 (INPUT) — search prompt not yet handled
- Image downloads not yet supported
