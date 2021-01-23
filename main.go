package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"

	"golang.org/x/net/html"
)

type Crawler struct {
	crawled map[string]bool
	mux     sync.Mutex
}

func New() *Crawler {
	return &Crawler{
		crawled: make(map[string]bool),
	}
}

func (c *Crawler) visit(url string) bool {
	c.mux.Lock()
	defer c.mux.Unlock()

	_, ok := c.crawled[url]
	if ok {
		return true
	}
	c.crawled[url] = true

	return false
}


func fetchUrls(url, baseUrl string, visited *map[string]bool) {
	response, err := http.Get(url)
	if err != nil {
		return
	}
	fmt.Println(response)
	page, err := html.Parse(response.Body)
	if err != nil {
		return
	}
	(*visited)[url] = true
	links := fetchLinks(nil, page, visited)
	for _, link := range links {
		if !(*visited)[link] && strings.HasPrefix(link, baseUrl) {
			fetchUrls(link, baseUrl, visited)
		}
	}

}

func fetchLinks(links []string, n *html.Node, visited *map[string]bool) []string {
	if n.Type == html.ElementNode && n.Data == "a" {
		for _, a := range n.Attr {
			if a.Key == "href" {
				_, ok := (*visited)[a.Val]
				if !ok {
					links = append(links, a.Val)
				}

				}
			}
		}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		links = fetchLinks(links, c, visited)
	}
	return links
}



func main() {
	url := os.Args[1]
	if url == "" {
		fmt.Println("Usage: `webcrawler <url>`")
		os.Exit(1)
	}

	visited := map[string]bool{}
	fetchUrls(url, url, &visited)
}