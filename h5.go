package main

import (
	"code.google.com/p/go.net/html"
	"flag"
	"io"
	"log"
	"os"
	"strings"
)

var (
	IndentWidth = flag.Int("w", 4, "Number of space or tab to indent")
	UseTab      = flag.Bool("t", false, "Use tab indent")
)

func Traverse(n *html.Node, callback func(*html.Node)) {
	callback(n)
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		Traverse(c, callback)
	}
}

func IsPartial(n *html.Node) bool {
	return n != nil &&
		n.FirstChild != nil && n.FirstChild.Data == "html" && // root's child is html
		n.FirstChild.NextSibling == nil && // only one element at html level
		n.FirstChild.FirstChild != nil && n.FirstChild.FirstChild.Data == "head" && // html only has head and body
		n.FirstChild.FirstChild.NextSibling != nil && n.FirstChild.FirstChild.NextSibling.Data == "body" &&
		n.FirstChild.FirstChild.NextSibling.NextSibling == nil &&
		(n.FirstChild.FirstChild.FirstChild == nil || n.FirstChild.FirstChild.NextSibling.FirstChild == nil) // head or body is empty
}

func Root(n *html.Node) *html.Node {
	if IsPartial(n) {
		head := n.FirstChild.FirstChild
		if head.FirstChild == nil {
			// body
			head.NextSibling.Type = html.DocumentNode
			return head.NextSibling
		} else {
			head.Type = html.DocumentNode
			return head
		}
	}
	return n
}

func SmartParse(r io.Reader) (*html.Node, error) {
	doc, err := html.Parse(os.Stdin)
	if err != nil {
		return doc, err
	}
	root := Root(doc)
	Clean(root)
	return root, err
}

func CleanNodeText(s string) string {
	ret := []string{}
	for _, line := range strings.FieldsFunc(s, func(r rune) bool { return r == '\n' || r == '\r' }) {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			ret = append(ret, trimmed)
		}
	}
	return strings.Join(ret, "\n")
}

func Clean(root *html.Node) {
	Traverse(root, func(n *html.Node) {
		toRemove := []*html.Node{}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			c.Data = CleanNodeText(c.Data)
			if c.Data == "" {
				toRemove = append(toRemove, c)
			}
		}
		for _, c := range toRemove {
			n.RemoveChild(c)
		}
	})
}

// taken from render
// Section 12.1.2, "Elements", gives this list of void elements. Void elements
// are those that can't have any contents.
var voidElements = map[string]bool{
	"area":    true,
	"base":    true,
	"br":      true,
	"col":     true,
	"command": true,
	"embed":   true,
	"hr":      true,
	"img":     true,
	"input":   true,
	"keygen":  true,
	"link":    true,
	"meta":    true,
	"param":   true,
	"source":  true,
	"track":   true,
	"wbr":     true,
}

func IsInline(n *html.Node) bool {
	return !strings.Contains(n.Data, "\n")
}

func IsTextNode(n *html.Node) bool {
	return n.Type == html.TextNode || n.Type == html.CommentNode
}

func IsInlineTextNode(n *html.Node) bool {
	return IsTextNode(n) && IsInline(n)
}

func ShouldIndent(n *html.Node) bool {
	if n == nil {
		return false
	}
	// inline text node
	if IsTextNode(n) && !IsInline(n) {
		return true
	}
	// no child, no indent
	if n.FirstChild == nil {
		return false
	}
	if n.FirstChild.NextSibling == nil {
		// only child is inline text node, no indent
		if IsInlineTextNode(n.FirstChild) {
			return false
		}
		// only grand child is inline text node, no indent
		if n.FirstChild.FirstChild != nil && n.FirstChild.FirstChild.NextSibling == nil &&
			IsInlineTextNode(n.FirstChild.FirstChild) {
			return false
		}
	}
	return true
}

func IndentPrint(n *html.Node, w io.Writer, indentWith string, width int, level int) {
	parentIndent := ShouldIndent(n.Parent)
	indent := ShouldIndent(n)
	space := strings.Repeat(indentWith, level*width)
	switch n.Type {
	case html.CommentNode:
		if parentIndent {
			io.WriteString(w, space)
		}
		io.WriteString(w, "<!--")
		if indent {
			io.WriteString(w, "\n")
		} else {
			io.WriteString(w, " ")
		}
		for _, line := range strings.Split(n.Data, "\n") {
			if indent {
				io.WriteString(w, space)
			}
			io.WriteString(w, line)
			if indent {
				io.WriteString(w, "\n")
			}
		}
		if indent {
			io.WriteString(w, space)
		} else {
			io.WriteString(w, " ")
		}
		io.WriteString(w, "-->")
		if parentIndent {
			io.WriteString(w, "\n")
		}
	case html.DoctypeNode:
		io.WriteString(w, "<!DOCTYPE html>\n")
	case html.TextNode:
		for _, line := range strings.Split(n.Data, "\n") {
			if parentIndent {
				io.WriteString(w, space)
			}
			io.WriteString(w, line)
			if parentIndent {
				io.WriteString(w, "\n")
			}
		}
	case html.ElementNode:
		if parentIndent {
			io.WriteString(w, space)
		}
		io.WriteString(w, "<"+n.Data)
		for _, a := range n.Attr {
			io.WriteString(w, " "+a.Key)
			if a.Val != "" {
				io.WriteString(w, "=\""+a.Val+"\"")
			}
		}
		if voidElements[n.Data] {
			io.WriteString(w, " />")
			if parentIndent {
				io.WriteString(w, "\n")
			}
			return
		}
		io.WriteString(w, ">")
		if indent {
			io.WriteString(w, "\n")
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if n.Data == "html" {
				IndentPrint(c, w, indentWith, width, level)
			} else {
				IndentPrint(c, w, indentWith, width, level+1)
			}
		}
		if indent {
			io.WriteString(w, space)
		}
		io.WriteString(w, "</"+n.Data+">")
		if parentIndent {
			io.WriteString(w, "\n")
		}
	case html.DocumentNode:
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			IndentPrint(c, w, indentWith, width, level)
		}
	}
}

func Render(n *html.Node, w io.Writer, indentWith string, width int) {
	IndentPrint(n, w, indentWith, width, 0)
}

func main() {
	flag.Parse()
	doc, err := SmartParse(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	indentString := " "
	if *UseTab {
		indentString = "\t"
	}
	if *IndentWidth <= 0 {
		log.Fatal("width should be positive")
	}
	Render(doc, os.Stdout, indentString, *IndentWidth)
	//html.Render(os.Stdout, doc)
}
