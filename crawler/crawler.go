package crawler

import (
	"fmt"
	"sync"
)

type Crawler struct {
	crawled map[string]bool
	mux     sync.Mutex
}

type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(url string) (body string, urls []string, err error)
}

type myFetcher map[string]*Result

type Result struct {
	body string
	urls []string
}

func (f myFetcher) Fetch(url string) (string, []string, error) {

	if res, ok := f[url]; ok {
		return res.body, res.urls, nil
	}
	return "", nil, fmt.Errorf("not found: %s", url)
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



// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func (c *Crawler) Crawl(url string, depth int, fetcher Fetcher) {
	var wg sync.WaitGroup

	v := c.visit(url)
	if v || depth <= 0 {
		return
	}

	body, urls, err := fetcher.Fetch(url)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("found: %s %q\n", url, body)

	for _, u := range urls {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()
			c.Crawl(u, depth-1, fetcher)
		}(u)
	}
	wg.Wait()
	return
}
