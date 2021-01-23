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



func (c *Crawler) fetchLinks(links []string, n *html.Node, baseUrl string) []string {
	if n.Type == html.ElementNode && n.Data == "a" {
		for _, a := range n.Attr {
			if a.Key == "href" && strings.HasPrefix(a.Val, baseUrl){
				c.visit(a.Val)
				}
			}
		}
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		links = c.fetchLinks(links, child, baseUrl)
	}
	return links
}

func (c *Crawler) Crawl(url string) {
	var wg sync.WaitGroup

	c.visit(url)

	response, err := http.Get(url)
	if err != nil {
		return
	}

	page, err := html.Parse(response.Body)
	if err != nil {
		return
	}

	urls := c.fetchLinks(nil, page, url)

	for _, u := range urls {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()
			c.Crawl(u)
		}(u)
	}
	wg.Wait()
	return
}

func main() {
	url := os.Args[1]
	if url == "" {
		fmt.Println("Usage: `webcrawler <url>`")
		os.Exit(1)
	}
		crawler := New()
		crawler.Crawl(url)
		fmt.Println(crawler.crawled)
	}
