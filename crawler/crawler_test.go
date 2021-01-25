package crawler

import (
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
		t.Errorf("%v does not have prefix %v", childUrl, baseUrl)
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