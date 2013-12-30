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
		n.FirstChild.FirstChild.FirstChild == nil // head is empty
}

func Root(n *html.Node) *html.Node {
	if IsPartial(n) {
		body := n.FirstChild.FirstChild.NextSibling
		body.Type = html.DocumentNode
		return body
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

func Clean(root *html.Node) {
	Traverse(root, func(n *html.Node) {
		toRemove := []*html.Node{}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			c.Data = strings.TrimSpace(c.Data)
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

func ShouldIndent(n *html.Node) bool {
	// no child, no indent
	if n.FirstChild == nil {
		return false
	}
	// indent script is probably a good idea
	if n.Data == "script" {
		return true
	}

	if n.FirstChild.NextSibling == nil {
		// only child is text node, no indent
		if n.FirstChild.Type == html.TextNode {
			return false
		} else {
			// only grand child is text node, no indent
			if n.FirstChild.FirstChild != nil && n.FirstChild.FirstChild.NextSibling == nil &&
				n.FirstChild.FirstChild.Type == html.TextNode {
				return false
			}
		}
	}
	return true
}

func IndentPrint(n *html.Node, w io.Writer, indentWith string, width int, level int) {
	switch n.Type {
	case html.CommentNode:
		if ShouldIndent(n.Parent) {
			io.WriteString(w, strings.Repeat(indentWith, level*width))
		}
		io.WriteString(w, "<!-- "+n.Data+" -->")
		if ShouldIndent(n.Parent) {
			io.WriteString(w, "\n")
		}
	case html.DoctypeNode:
		io.WriteString(w, "<!DOCTYPE html>\n")
	case html.TextNode:
		if ShouldIndent(n.Parent) {
			io.WriteString(w, strings.Repeat(indentWith, level*width))
		}
		io.WriteString(w, n.Data)
		if ShouldIndent(n.Parent) {
			io.WriteString(w, "\n")
		}

	case html.ElementNode:
		if ShouldIndent(n.Parent) {
			io.WriteString(w, strings.Repeat(indentWith, level*width))
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
			if ShouldIndent(n.Parent) {
				io.WriteString(w, "\n")
			}
			return
		}
		io.WriteString(w, ">")
		// script is exception as we'll always indent
		if ShouldIndent(n) {
			io.WriteString(w, "\n")
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if n.Data == "html" {
				IndentPrint(c, w, indentWith, width, level)
			} else {
				IndentPrint(c, w, indentWith, width, level+1)
			}
		}
		if ShouldIndent(n) {
			io.WriteString(w, strings.Repeat(indentWith, level*width))
		}
		io.WriteString(w, "</"+n.Data+">")
		if ShouldIndent(n.Parent) {
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
