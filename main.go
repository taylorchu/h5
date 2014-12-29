package main

import (
	"log"
	"os"
	"strings"

	. "html"

	"github.com/taylorchu/h5/pretty"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type Node struct {
	*html.Node
}

func (n *Node) Children() []pretty.Node {
	var ns []pretty.Node
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		ns = append(ns, &Node{c})
	}
	return ns
}

func (n *Node) Start() string {
	switch n.Type {
	case html.CommentNode:
		return "<!--"
	case html.DoctypeNode:
		return "<!DOCTYPE html>"
	case html.ElementNode:
		s := "<" + n.Data
		for _, a := range n.Attr {
			s += " " + a.Key
			if a.Val != "" {
				s += `="` + EscapeString(a.Val) + `"`
			}
		}
		if voidElements[n.Data] {
			s += " />"
		} else {
			s += ">"
		}
		return s
	default:
		return ""
	}
}

func (n *Node) End() string {
	switch n.Type {
	case html.CommentNode:
		return "-->"
	case html.ElementNode:
		if voidElements[n.Data] {
			return ""
		}
		return "</" + n.Data + ">"
	default:
		return ""
	}
}

func (n *Node) Parent() pretty.Node {
	if n.Node.Parent == nil {
		return nil
	}
	return &Node{n.Node.Parent}
}

func (n *Node) Text() []string {
	switch n.Type {
	case html.CommentNode:
		fallthrough
	case html.TextNode:
		s := n.Data
		if n.Node.Parent != nil {
			if n.Node.Parent.DataAtom != atom.Script {
				s = EscapeString(s)
			}
			if n.Node.Parent.DataAtom == atom.Pre {
				return []string{s}
			}
		}
		return strings.Split(strings.Trim(s, " \t\r\n"), "\n")
	default:
		return nil
	}
}

func (n *Node) Inline() bool {
	switch n.DataAtom {
	case atom.Pre:
		return true
	}
	if n.Type == html.ElementNode && n.FirstChild == nil {
		return true
	}
	if n.FirstChild != nil && n.FirstChild.NextSibling == nil &&
		n.FirstChild.Type == html.TextNode &&
		!strings.ContainsAny(n.FirstChild.Data, "\r\n") {
		return true
	}
	return n.Parent() != nil && n.Parent().Inline()
}

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

func clean(n *html.Node) {
	for c := n.FirstChild; c != nil; {
		// c.NextSibling is nil once it is removed
		tmp := c.NextSibling
		if strings.Trim(c.Data, " \t\r\n") == "" {
			n.RemoveChild(c)
		}
		c = tmp
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		clean(c)
	}
}

func main() {
	doc, err := html.Parse(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	clean(doc)
	pretty.Print(os.Stdout, &Node{doc})
}
