package crawler

import (
	"golang.org/x/net/html"
	"os"
	"strings"
	"testing"
)

func TestIsChildUrl(t *testing.T) {
	baseUrl := "https://www.my_test.com"
	childUrl := "https://www.my_test.com/another_test"
	otherUrl := "https://www.other_test.com"
	if !IsUrlWithBase(childUrl, baseUrl) {
		t.Errorf("%v has prefix %v", childUrl, baseUrl)
	}
	if IsUrlWithBase(otherUrl, baseUrl) {
		t.Errorf("%v does not have prefix %v", otherUrl, baseUrl)
	}
}

func TestFetchLinks(t *testing.T) {
	file, err := os.Open("example.html")
	if err != nil {
		t.Errorf("Failed to load local html file")
	}
	doc, err := html.Parse(file)
	crawler := NewCrawler(".")
	crawler.Writer.OpenFile()
	baseUrl := "https://www.my_test.com"
	defer crawler.Writer.CloseFile()
	links := crawler.FetchLinks(nil, doc, baseUrl)
	if len(links) != 2 {
		t.Errorf("Wrong amount of links")
	}
	if !allLinksStartWith(links, baseUrl) {
		t.Errorf("Not all links starts with %v", baseUrl)
	}

}

func allLinksStartWith(links []string, prefix string) bool {
	for i := 0; i < len(links); i++ {
		if !strings.HasPrefix(links[i], prefix) {
			return false
		}
	}
	return true
}