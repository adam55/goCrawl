package main

import (
	"fmt"
	"golang.org/x/net/html"
	"net/http"
	"os"
	"strings"
	"sync"
)

type Crawler struct {
	crawled map[string]bool
	mux     sync.Mutex
	writer Writer
}

type Writer struct {
	filePath string
	mux sync.Mutex
}

func NewWriter(filePath string) *Writer {
	// open output file
	file, err := os.Create(filePath)
	if err != nil {
		panic(err)
	}
	// close fo on exit and check for its returned error
	defer func() {
		if err := file.Close(); err != nil {
			panic(err)
		}
	}()
	return &Writer{
		filePath: filePath,
	}
}

func NewCrawler(directory string) *Crawler {
	return &Crawler{
		crawled: make(map[string]bool),
		writer: *NewWriter(directory + "/output.txt"),
	}
}

func (c *Crawler) visit(url string) bool {
	c.mux.Lock()
	defer c.mux.Unlock()
	_, ok := c.crawled[url]
	if ok {
		return true
	}
	fmt.Println(url)
	c.crawled[url] = true
	return false
}

func (w *Writer) write(url string, file *os.File) {
	w.mux.Lock()
	defer w.mux.Unlock()
	_, err := file.WriteString(url)
	if err != nil {
		fmt.Println(err)
	}
}

func (w *Writer) writeBatch(urls []string, wg *sync.WaitGroup) {
	// open input file
	file, err := os.Open(w.filePath)
	if err != nil {
		panic(err)
	}
	// close fi on exit and check for its returned error
	defer func() {
		if err := file.Close(); err != nil {
			panic(err)
		}
	}()
	for _, u := range urls {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()
			w.write(u, file)
		}(u)
	}
}


func (c *Crawler) fetchLinks(links []string, n *html.Node, baseUrl string) []string {
	if n.Type == html.ElementNode && n.Data == "a" {
		for _, a := range n.Attr {
			if a.Key == "href" && strings.HasPrefix(a.Val, baseUrl){
				if !c.visit(a.Val) {
					fmt.Println(a.Val)
				}
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
		fmt.Println("failed to get url response")
		return
	}

	page, err := html.Parse(response.Body)
	if err != nil {
		fmt.Println("failed to parse response's body")
		return
	}

	urls := c.fetchLinks(nil, page, url)
	c.writer.writeBatch(urls, &wg)

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
		fmt.Println("Usage: `webcrawler <url> <target_directory>`")
		os.Exit(1)
	}

	targetDirectory := os.Args[2]
	if targetDirectory == "" {
		fmt.Println("Usage: `webcrawler <url> <target_directory>")
		os.Exit(1)
	}
		crawler := NewCrawler(targetDirectory)
		crawler.Crawl(url)
	}
