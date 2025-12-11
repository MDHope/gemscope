package ui

import "github.com/MDHope/gemscope/internal/gemtext"

type LinkHint struct {
	Keys   string
	Node   *gemtext.Node
	Line   int
	Column int
}
