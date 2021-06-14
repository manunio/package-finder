package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	fmt.Println("Starting service..")

	// var urls []string
	// var allSources []string

	// read urls
	urls, err := readFile("urls.txt")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("read lines:")
	for _, url := range urls {
		fmt.Println(url)
	}

	return

}

func readFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	// buffer
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, nil
}

func getScriptSrc(url, method string, headers []string, insecure bool, timeout int) ([]string, error) {
	// req the HTML page
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return []string{}, err
	}
	for _, d := range headers {
		values := strings.Split(d, ":")
		if len(values) == 2 {
			fmt.Println("[+] New Header: " + values[0] + ": " + values[1])
			req.Header.Set(values[0], values[1])
		}
	}

	tr := &http.Transport{
		ResponseHeaderTimeout: time.Duration(time.Duration(timeout) * time.Second),
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: insecure},
	}

	var client = &http.Client{
		Timeout:   time.Duration(time.Duration(timeout) * time.Second),
		Transport: tr,
	}

	res, err := client.Do(req)
	if err != nil {
		return []string{}, err
	}

	defer res.Body.Close()
	if res.StatusCode != 200 {
		fmt.Printf("[!]%s returned an %d instead of %d", url, res.StatusCode, http.StatusOK)
		return nil, nil
	}

	// Load the html document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	var sources []string

	// Find the script tag and get the src
	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		src, _ := s.Attr("src")
		dsrc, _ := s.Attr("data-src")
		if src != "" {
			sources = append(sources, src)
		}
		if dsrc != "" {
			sources = append(sources, dsrc)
		}
	})

	return sources, nil

}
