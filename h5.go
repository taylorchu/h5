package main

import (
	"code.google.com/p/go.net/html"
	"io"
	"log"
	"os"
	"strings"
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
			if c.Type == html.TextNode {
				c.Data = strings.TrimSpace(c.Data)
				if c.Data == "" {
					toRemove = append(toRemove, c)
				}
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
	// has more than one node, or is not text node
	// or its only grandchildren is not text node
	return n.FirstChild != nil &&
		(n.FirstChild.Type != html.TextNode || n.FirstChild.NextSibling != nil)
}

func IndentPrint(n *html.Node, w io.Writer, indentWith string, width int, level int) {
	switch n.Type {
	case html.CommentNode:
		io.WriteString(w, strings.Repeat(indentWith, level*width) + "<!--" + n.Data + "-->\n")
	case html.DoctypeNode:
		io.WriteString(w, "<!DOCTYPE html>\n")
	case html.TextNode:
		alone := true
		// script is an exception
		if n.PrevSibling == nil && n.NextSibling == nil && n.Parent.Data != "script" {
			alone = false
		}
		if alone {
			io.WriteString(w, strings.Repeat(indentWith, level*width))
		}
		io.WriteString(w, n.Data)
		if alone {
			io.WriteString(w, "\n")
		}

	case html.ElementNode:
		io.WriteString(w, strings.Repeat(indentWith, level*width))
		io.WriteString(w, "<"+n.Data)
		for _, a := range n.Attr {
			io.WriteString(w, " "+a.Key+"=\""+a.Val+"\"")
		}
		if voidElements[n.Data] {
			io.WriteString(w, "/>\n")
			return
		}
		io.WriteString(w, ">")
		// script is exception as we'll always indent
		if ShouldIndent(n) || (n.Data == "script" && n.FirstChild != nil) {
			io.WriteString(w, "\n")
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if n.Data == "html" {
				IndentPrint(c, w, indentWith, width, level)
			} else {
				IndentPrint(c, w, indentWith, width, level+1)
			}
		}
		if ShouldIndent(n) || (n.Data == "script" && n.FirstChild != nil) {
			io.WriteString(w, strings.Repeat(indentWith, level*width))
		}
		io.WriteString(w, "</"+n.Data+">\n")
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
	doc, err := SmartParse(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	Render(doc, os.Stdout, " ", 4)
	//html.Render(os.Stdout, doc)
}
