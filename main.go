package main

import (
	"fmt"
	"os"
	"time"
	crawler2 "webCrawler/crawler"
)



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

	startTime := time.Now()
	crawler := crawler2.NewCrawler(targetDirectory)
	crawler.Writer.OpenFile()
	defer crawler.Writer.CloseFile()
	crawler.Crawl(url, url)
	defer fmt.Printf("Running Time: %v sec", time.Now().Sub(startTime).Seconds())

}