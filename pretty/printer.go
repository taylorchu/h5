package pretty

import (
	"io"
	"strings"
)

var (
	IndentString = "\t"
)

type Node interface {
	Inline() bool
	Start() string
	Text() []string
	End() string
	Parent() Node
	Children() []Node
}

func print(w io.Writer, n Node, depth int) {
	repeat := strings.Repeat(IndentString, depth)
	repeat2 := strings.Repeat(IndentString, depth+1)
	lines := n.Text()
	inline := n.Inline()
	start := n.Start()
	end := n.End()
	children := n.Children()

	if start != "" {
		if !inline {
			io.WriteString(w, repeat)
		}
		io.WriteString(w, start)
		if !inline {
			io.WriteString(w, "\n")
		}
	}

	for _, line := range lines {
		if !inline {
			if start != "" {
				io.WriteString(w, repeat2)
			} else {
				io.WriteString(w, repeat)
			}
		}
		io.WriteString(w, line)
		if !inline {
			io.WriteString(w, "\n")
		}
	}

	for _, c := range children {
		if c.Inline() && !inline {
			io.WriteString(w, repeat2)
		}
		if start == "" && end == "" && len(lines) == 0 {
			print(w, c, depth)
		} else {
			print(w, c, depth+1)
		}
		if c.Inline() && !inline {
			io.WriteString(w, "\n")
		}
	}

	if end != "" {
		if !inline {
			io.WriteString(w, repeat)
		}
		io.WriteString(w, end)
		if !inline {
			io.WriteString(w, "\n")
		}
	}
}

func Print(w io.Writer, n Node) {
	print(w, n, 0)
}
