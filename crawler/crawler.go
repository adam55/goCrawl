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

func IsUrlWithBase(url string, baseUrl string) bool {
	return strings.HasPrefix(url, baseUrl)
}

func LeadsToChildUrl(hrefValue string) bool {
	//todo refactor this into something less hacky
	return strings.HasPrefix(hrefValue, "/") && len(hrefValue)>= 2
}

func preprocessUrl(baseUrl string) string {
	if baseUrl[len(baseUrl) - 1] == '/' {
		return baseUrl[:len(baseUrl) - 1]
	}
	return baseUrl
}
func (c *Crawler) ProcessNodeAttribute(a *html.Attribute, baseUrl string, wg *sync.WaitGroup) {
	if a.Key == "href" {
		if IsUrlWithBase(a.Val, baseUrl) {
			preprocessedUrl := preprocessUrl(a.Val)
			if !c.Visit(preprocessedUrl) {
				c.Writer.Write(preprocessedUrl)
				wg.Add(1)
				go func(u string) {
					defer wg.Done()
					c.Crawl(u)
				}(preprocessedUrl)
			}
		}
		if  LeadsToChildUrl(a.Val){
			reconstructedUrl := preprocessUrl(baseUrl + a.Val)
			if !c.Visit(reconstructedUrl) {
				c.Writer.Write(reconstructedUrl + "\n")
				wg.Add(1)
				go func(u string) {
					defer wg.Done()
					c.Crawl(u)
				}(reconstructedUrl)
			}
		}
	}
}

func (c *Crawler) FetchLinks(n *html.Node, baseUrl string, wg *sync.WaitGroup) {
	if n.Type == html.ElementNode && n.Data == "a" {
		for _, a := range n.Attr {
			c.ProcessNodeAttribute(&a, baseUrl, wg)
		}
	}
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		c.FetchLinks(child, baseUrl, wg)
	}
}

func (c *Crawler) Crawl(url string) {
	var wg sync.WaitGroup
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
	baseUrl := preprocessUrl(url)
	if !c.Visit(baseUrl) {
		c.Writer.Write(baseUrl)
	}
	c.FetchLinks(page, baseUrl, &wg)
	wg.Wait()
}
