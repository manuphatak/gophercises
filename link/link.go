package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

type Link struct {
	Href string
	Text string
}

func main() {
	useHtml := flag.Bool("html", false, "Get links from a remote HTML file")

	flag.Parse()

	target := flag.Arg(0)
	if target == "" {
		fmt.Println("usage: link <target>")
		os.Exit(1)
	}

	var reader io.Reader

	if *useHtml {
		response, err := http.Get(target)
		check(err)
		reader = response.Body
	} else {
		file, err := os.Open(target)
		check(err)
		defer file.Close()
		reader = file
	}

	links := extractLinks(reader)

	for _, link := range links {
		fmt.Printf("%#v\n", link)
	}
}

func extractLinks(reader io.Reader) []Link {
	doc, err := html.Parse(reader)
	check(err)

	links := []Link{}
	walk(doc, &links)

	return links
}

func walk(node *html.Node, links *[]Link) {
	switch node.Type {
	case html.ElementNode:
		if node.Data == "a" {
			for _, attr := range node.Attr {
				if attr.Key == "href" {
					*links = append(*links, Link{attr.Val, findLinkText(node, 0)})
					return
				}
			}
		}
	}

	for next := node.FirstChild; next != nil; next = next.NextSibling {
		walk(next, links)
	}
}

func findLinkText(node *html.Node, depth int32) string {
	text := ""
	re := regexp.MustCompile(`(?m)\s+`)

	if next := node.FirstChild; next != nil {
		if next.Type == html.TextNode {
			text = text + next.Data + findLinkText(next, depth+1)
		} else {
			text = text + findLinkText(next, depth+1)
		}
	}

	if next := node.NextSibling; next != nil && depth > 0 {
		if next.Type == html.TextNode {
			text = text + next.Data + findLinkText(next, depth+1)
		} else {
			text = text + findLinkText(next, depth+1)
		}
	}

	return strings.TrimSpace(re.ReplaceAllString(text, " "))
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
