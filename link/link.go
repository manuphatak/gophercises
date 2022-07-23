package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"golang.org/x/net/html"
)

type Link struct {
	Href string
	Text string
}

func main() {
	flag.Parse()
	site := flag.Arg(0)
	if site == "" {
		fmt.Println("usage: link <target>")
		os.Exit(1)
	}

	links := extractLinks(site)

	for _, link := range links {
		fmt.Printf("%#v\n", link)
	}
}

func extractLinks(fileName string) []Link {
	file, err := os.Open(fileName)
	check(err)
	defer file.Close()

	doc, err := html.Parse(file)
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

	if next := node.FirstChild; next != nil {
		if next.Type == html.TextNode {
			text = text + strings.TrimSpace(next.Data) + findLinkText(next, depth+1)
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

	return strings.TrimSpace(text)
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
