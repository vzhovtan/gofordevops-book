package cli

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"golang.org/x/net/html"
)

func Run(fname string) {
	text, err := readFile(fname)
	if err != nil {
		log.Fatal(err)
	}

	doc, err := html.Parse(strings.NewReader(text))
	if err != nil {
		fmt.Fprintf(os.Stderr, "outline: %v\n", err)
		os.Exit(1)
	}

	outline(os.Stdout, nil, doc)
}

func readFile(fname string) (string, error) {
	bs, err := os.ReadFile(fname)
	if err != nil {
		return "", err
	}
	return string(bs), nil
}

func outline(w io.Writer, stack []string, n *html.Node) {
	if n.Type == html.ElementNode {
		stack = append(stack, n.Data) // push tag
		fmt.Fprintf(w, "%v\n", stack)
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		outline(w, stack, c)
	}
}
