package gemtext

import (
	"fmt"
	"strings"
)

type Tag string

const (
	Heading1Tag Tag = "#"
	Heading2Tag Tag = "##"
	Heading3Tag Tag = "###"

	ListItemTag Tag = "*"

	BlockquoteTag Tag = ">"

	LinkTag Tag = "=>"

	PreformattedTag Tag = "```"
)

type NodeType int

const (
	NodeDocument NodeType = iota
	NodeHeading1
	NodeHeading2
	NodeHeading3
	NodeText
	NodeLink
	NodePreformatted
	NodeBlockquote
	NodeListItem
	NodeBlankLine NodeType = iota + 100
)

type Node struct {
	Type       NodeType
	RawContent string
	Children   []*Node
	Parent     *Node
	Meta       any
}

type LinkMeta struct {
	Url   string
	Label string
	// IsGeminiLink bool
}

type PreformattedMeta struct {
	Content string
}

type TextMeta struct {
	Content string
}

func Parse(input string) *Node {
	root := &Node{Type: NodeDocument}

	stack := []*Node{root, nil, nil, nil}

	lines := strings.Split(input, "\n")
	var inPreformatted bool
	var preNode *Node

	for _, line := range lines {
		if strings.HasPrefix(line, string(PreformattedTag)) {
			if !inPreformatted {
				inPreformatted = true
				alt := strings.TrimPrefix(line, string(PreformattedTag))
				preNode = &Node{Type: NodePreformatted, Meta: PreformattedMeta{Content: alt}}
				attachNode(stack, preNode, 4)
			} else {
				inPreformatted = false
				preNode = nil
			}
			continue
		}

		if inPreformatted {
			preNode.RawContent += line + "\n"
			continue
		}

		node, level := parseLine(line)
		if node != nil {
			attachNode(stack, node, level)
		}
	}

	return root
}

func attachNode(stack []*Node, node *Node, level int) {
	if level >= 1 && level <= 3 {
		for i := level + 1; i < len(stack); i++ {
			stack[i] = nil
		}
		stack[level] = node
	}

	var parent *Node
	for i := level - 1; i >= 0; i-- {
		if stack[i] != nil {
			parent = stack[i]
			break
		}
	}

	node.Parent = parent
	parent.Children = append(parent.Children, node)
}

func parseLine(line string) (*Node, int) {
	switch {
	case strings.HasPrefix(line, string(Heading3Tag)):
		return &Node{Type: NodeHeading3, RawContent: line, Meta: TextMeta{Content: line[4:]}}, 3
	case strings.HasPrefix(line, string(Heading2Tag)):
		return &Node{Type: NodeHeading2, RawContent: line, Meta: TextMeta{Content: line[3:]}}, 2
	case strings.HasPrefix(line, string(Heading1Tag)):
		return &Node{Type: NodeHeading1, RawContent: line, Meta: TextMeta{Content: line[2:]}}, 1
	case strings.HasPrefix(line, string(LinkTag)):
		return parseLink(line), 4
	case strings.HasPrefix(line, string(BlockquoteTag)):
		return &Node{Type: NodeBlockquote, RawContent: line, Meta: TextMeta{Content: strings.TrimPrefix(line, ">")}}, 4
	case strings.HasPrefix(line, string(LinkTag)):
		content := strings.TrimPrefix(line, "*")
		content = strings.TrimSpace(content)
		return &Node{Type: NodeListItem, RawContent: line, Meta: TextMeta{Content: content}}, 4
	case strings.TrimSpace(line) == "":
		return &Node{Type: NodeBlankLine}, 4
	case strings.TrimSpace(line) != "":
		return &Node{Type: NodeText, RawContent: line, Meta: TextMeta{Content: line}}, 4
	}

	return nil, 0
}

func parseLink(line string) *Node {
	trimmedLine := strings.TrimPrefix(line, "=>")
	trimmedLine = strings.TrimSpace(trimmedLine)

	parts := strings.Fields(trimmedLine)
	url := strings.TrimSpace(parts[0])
	label := ""
	if len(parts) > 1 {
		label = strings.TrimSpace(strings.Join(parts[1:], " "))
	}
	node := &Node{Type: NodeLink, RawContent: line, Meta: LinkMeta{Url: url, Label: label}}

	return node
}

func (n *Node) Print() string {
	var sb strings.Builder
	n.print(&sb, 0)
	return sb.String()
}

func (n *Node) print(sb *strings.Builder, depth int) {
	indent := strings.Repeat(" ", depth)
	sb.WriteString(indent)
	sb.WriteString(n.typeName())

	var content string

	switch n.Meta.(type) {
	case TextMeta:
		content = n.Meta.(TextMeta).Content
		break
	case LinkMeta:
		content = n.Meta.(LinkMeta).Url
		break
	case PreformattedMeta:
		content = n.Meta.(PreformattedMeta).Content
		break
	}

	if content != "" {
		cont := content
		if len(cont) > 50 {
			cont = cont[:47] + "..."
		}

		cont = strings.ReplaceAll(cont, "\n", "\\n")
		sb.WriteString(fmt.Sprintf(" %q", cont))
	}

	if meta, ok := n.Meta.(LinkMeta); ok && meta.Label != "" {
		sb.WriteString(fmt.Sprintf(" label=%q", n.Meta.(LinkMeta).Label))
	}

	sb.WriteString("\n")

	for _, child := range n.Children {
		child.print(sb, depth+1)
	}
}

func (n *Node) typeName() string {
	names := []string{"Document", "H1", "H2", "H3", "Text", "Link", "Pre", "Quote", "ListItem"}

	if int(n.Type) < len(names) {
		return names[n.Type]
	}

	return fmt.Sprintf("Unknown(%d)", n.Type)
}
