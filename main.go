package main

import (
	"fmt"
	"os"
	"os/signal"
	_ "os/signal"
	"syscall"
	_ "syscall"
	"time"
	crawler2 "webCrawler/crawler"
)


func SetupCloseHandler(writer *crawler2.Writer) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\r- Ctrl+C pressed in Terminal")
		writer.CloseFile()
		fmt.Printf("Output file closed")
		os.Exit(0)
	}()
}

func main() {

	u := os.Args[1]
	if u == "" {
		fmt.Println("Usage: `webcrawler <url> <target_directory>`")
		os.Exit(1)
	}

	targetDirectory := os.Args[2]
	if targetDirectory == "" {
		fmt.Println("Usage: `webcrawler <url> <target_directory>")
		os.Exit(1)
	}
	preprocessedUrl := crawler2.PreprocessUrl(u)
	baseUrl := crawler2.GetBaseUrl(preprocessedUrl)

	startTime := time.Now()
	crawler := crawler2.NewCrawler(targetDirectory)
	SetupCloseHandler(&crawler.Writer)
	crawler.Writer.OpenFile()
	defer crawler.Writer.CloseFile()
	fmt.Printf("Crawling from %v", preprocessedUrl)
	crawler.Crawl(preprocessedUrl, baseUrl, preprocessedUrl)
	defer fmt.Printf("Running Time: %v sec", time.Now().Sub(startTime).Seconds())

}