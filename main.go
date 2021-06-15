package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	p "path"
	"path/filepath"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// TODO: UNDERSTAND LOGIC PROPERLY
func main() {
	fmt.Println("Starting service..")

	// var urls []string
	// var allSources []string

	// read urls
	urls, err := readFile("test-urls.txt")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("read lines:")
	// getting js source file
	for _, u := range urls {
		fmt.Println("getting sources:")

		path, _ := standardizeURLForDirectoryName(u)
		if err := createDirector(path); err != nil {
			fmt.Println(err)
			return
		}

		sources, err := getScriptSrc(u, "GET", nil, false, 10)
		if err != nil {
			fmt.Println(err)
		}

		fmt.Println("printing sources:")

		for _, source := range sources {
			filenameURL, err := url.Parse(source)
			if err != nil {
				fmt.Println(err)
			}
			filename := p.Base(filenameURL.Path)
			if filename == "." {
				break
			}
			fullpath := filepath.Join("out", path, filename)
			if strings.HasPrefix(source, "/") {
				source = u + "/" + filename
				fmt.Println(fullpath)
			}
			if err := downloadFile(fullpath, source); err != nil {
				fmt.Println(err)
				return
			}
		}

	}

	return

}

// downloadFile downloads the given url and returns it,
// to be used by other parts of program
func downloadFile(path, url string) error {
	// get the data
	res, err := http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// create the file
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	// write the body to file
	_, err = io.Copy(out, res.Body)
	return err
}

// standardizeURLForDirectoryName extracts hostname from given link/url,
// then replaces '.' to '-' of hostname and returns it.
func standardizeURLForDirectoryName(link string) (string, error) {
	// extract hostname
	u, err := url.Parse(link)
	if err != nil {
		return "", err
	}
	// replaces '.' with  '-'
	hostname := u.Hostname()
	hostname = strings.Join(strings.Split(hostname, "."), "-")
	return hostname, nil
}

// createDirectory for creating directory based on provided name,
// in our case its url name
func createDirector(path string) error {
	// create an out directory if it doesn't already exists
	outputDirectory := "out"
	_, err := os.Stat(outputDirectory)
	if err != nil {
		if os.IsNotExist(err) {
			if err = os.Mkdir(outputDirectory, 0755); err != nil {
				return err
			}
		}
	}

	hostSubDir := filepath.Join(outputDirectory, path)
	if err := os.Mkdir(hostSubDir, 0755); err != nil {
		if os.IsExist(err) {
			return nil
		} else {
			return err
		}
	}

	return nil
}

// reaFile reads the given filename from disk line by line,
// adds it to a list of string and returns it.
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

// getScriptSrc gets javascript source file based on the url and method provided,
// source file is fetched based on src or data-src tag,
// from the queried document and a list of src urls is returned as a result.
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
