package crawler

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
	Writer  Writer
}

type Writer struct {
	filePath string
	mux      sync.Mutex
	file     *os.File
}

func (w *Writer) OpenFile() {
	w.mux.Lock()
	defer w.mux.Unlock()
	file, err := os.OpenFile(w.filePath,  os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		panic(err)
	}
	w.file = file
}

func (w *Writer) CloseFile() {
	w.mux.Lock()
	defer w.mux.Unlock()
	defer func() {
		if err := w.file.Close(); err != nil {
			panic(err)
		}
	}()
}
func NewWriter(filePath string) *Writer {
	file, err := os.Create(filePath)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			panic(err)
		}
	}()
	return &Writer{
		filePath: filePath,
		file: file,
	}
}

func NewCrawler(directory string) *Crawler {
	return &Crawler{
		crawled: make(map[string]bool),
		Writer:  *NewWriter(directory + "/output.txt"),
	}
}

func (c *Crawler) Visit(url string) bool {
	c.mux.Lock()
	defer c.mux.Unlock()
	_, ok := c.crawled[url]
	if ok {
		return true
	}
	c.crawled[url] = true
	fmt.Println(url)
	return false
}

func (w *Writer) Write(url string) {
	// open input file
	w.mux.Lock()
	defer w.mux.Unlock()
	_, err1 := w.file.WriteString(url)
	if err1 != nil {
		panic(err1)
	}
}



func (c *Crawler) FetchLinks(links []string, n *html.Node, baseUrl string) []string {
	if n.Type == html.ElementNode && n.Data == "a" {
		for _, a := range n.Attr {
			if a.Key == "href" {
				if strings.HasPrefix(a.Val, baseUrl) {
					if !c.Visit(a.Val) {
						c.Writer.Write(a.Val)
						links = append(links, a.Val)
					}
				}
				if  strings.HasPrefix(a.Val, "/") && len(a.Val)>= 2 {
					if !c.Visit(baseUrl + a.Val) {
						c.Writer.Write(baseUrl + a.Val + "\n")
						links = append(links, baseUrl + a.Val )
					}
				}
			}
		}
	}
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		links = c.FetchLinks(links, child, baseUrl)
	}
	return links
}

func (c *Crawler) Crawl(url string) {
	var wg sync.WaitGroup

	if !c.Visit(url) {
		c.Writer.Write(url)
	}
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

	urls := c.FetchLinks(nil, page, url)

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
