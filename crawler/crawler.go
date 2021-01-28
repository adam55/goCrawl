package crawler

import (
	"bytes"
	"fmt"
	"github.com/hashicorp/go-retryablehttp"
	"golang.org/x/net/html"
	"io"
	"net/url"
	_ "net/url"
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

func GetBaseUrl(u string) string{
	parsedUrl, err := url.Parse(u)
	if err != nil {
		panic(err)
	}
	return parsedUrl.Scheme + "://" + parsedUrl.Host
}
func (w *Writer) OpenFile() {
	file, err := os.OpenFile(w.filePath,  os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		panic(err)
	}
	w.file = file
}

func renderNode(n *html.Node) string {
	var buf bytes.Buffer
	w := io.Writer(&buf)
	err := html.Render(w, n)
	if err != nil {
		fmt.Println(err)
	}
	return buf.String()
}

func (w *Writer) CloseFile() {
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
	return false
}

func (w *Writer) Write(page *html.Node) {
	// open input file
	w.mux.Lock()
	defer w.mux.Unlock()
	data := renderNode(page)
	_, err1 := w.file.WriteString(data)
	if err1 != nil {
		panic(err1)
	}
}

func IsUrlWithBase(url string, baseUrl string) bool {
	return strings.HasPrefix(url, baseUrl)
}

func isSubPath(hrefValue string) bool {
	//todo refactor this into something less hacky
	return strings.HasPrefix(hrefValue, "/") && len(hrefValue)>= 2
}


func NodeLeadsToChild(a * html.Attribute, baseUrl string, inputUrl string) bool {
	return isSubPath(a.Val) && strings.HasPrefix(a.Val, inputUrl[len(baseUrl):])
}


func PreprocessUrl(baseUrl string) string {
	if baseUrl[len(baseUrl) - 1] == '/' {
		return baseUrl[:len(baseUrl) - 1]
	}
	return baseUrl
}
func (c *Crawler) ProcessNodeAttribute(a *html.Attribute, baseUrl string, inputUrl string, wg *sync.WaitGroup) {
	if a.Key == "href" {
		if IsUrlWithBase(a.Val, inputUrl) {
			preprocessedUrl := PreprocessUrl(a.Val)
			if !c.Visit(preprocessedUrl) {
				wg.Add(1)
				go func(u string) {
					defer wg.Done()
					c.Crawl(u, baseUrl, inputUrl)
				}(preprocessedUrl)
			}
		}
		if NodeLeadsToChild(a, baseUrl, inputUrl) {
			reconstructedUrl := PreprocessUrl(baseUrl + a.Val)
			if !c.Visit(reconstructedUrl) {
				wg.Add(1)
				go func(u string) {
					defer wg.Done()
					c.Crawl(u, baseUrl, inputUrl)
				}(reconstructedUrl)
			}
		}
	}
}

func (c *Crawler) FetchLinks(n *html.Node,baseUrl string, inputUrl string, wg *sync.WaitGroup) {
	if n.Type == html.ElementNode && n.Data == "a" {
		for _, a := range n.Attr {
			c.ProcessNodeAttribute(&a, baseUrl, inputUrl, wg)
		}
	}
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		c.FetchLinks(child, baseUrl, inputUrl, wg)
	}
}

func (c *Crawler) Crawl(url string, baseUrl string, inputUrl string) {
	var wg sync.WaitGroup
	response, err := retryablehttp.Get(url)
	if err != nil {
		fmt.Println(err)
		return
	}
	page, err := html.Parse(response.Body)
	if err != nil {
		fmt.Println("failed to parse response's body")
		return
	}
	c.Visit(url)
	c.Writer.Write(page)
	c.FetchLinks(page, baseUrl, inputUrl, &wg)
	wg.Wait()
}
