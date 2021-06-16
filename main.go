package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	pa "path"
	"path/filepath"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// package-finder downloads all javascript files from the given url,
// stores it, and reads it for private package.json file traces i.e,
// in our case it searches for "scripts:" in file and then logs useful details to a package.log
// in-case its found.

type Domain struct {
	Name     string `yaml:"name"`
	UrlsFile string `yaml:"urls_file"`
}

type Package struct {
	Domains    []Domain `yaml:"domains"`
	ConfigFile []byte
}

var (
	domainName string
)

func main() {
	log.Println("Starting service..")

	var domainUrlFile string

	p := Package{}
	if err := p.readConfigFile(); err != nil {
		log.Println("[ERROR] readConfigFile: " + err.Error())
		return
	}

	if p.Domains != nil {
		for _, d := range p.Domains {
			if d.Name != "" && d.UrlsFile != "" {
				domainUrlFile = d.UrlsFile
				domainName = d.Name
			}
		}
	}
	// read urls
	urls, err := readFile(domainUrlFile)
	if err != nil {
		log.Println("[ERROR] readFile: " + err.Error())
		return
	}
	// getting js source file
	for _, u := range urls {
		path, err := standardizeURLForDirectoryName(u)
		if err != nil {
			log.Println("[ERROR] standardizeURLForDirectoryName: " + err.Error())
			return
		}

		if err := createDirector(path); err != nil {
			log.Println("[ERROR] createDirector: " + err.Error())
			return
		}

		sources, err := getScriptSrc(u, "GET", nil, true, 10)
		if err != nil {
			log.Println("[ERROR] getScriptSrc: " + err.Error())
		}

		for _, source := range sources {
			log.Println(source)
			filenameURL, err := url.Parse(source)
			if err != nil {
				log.Println("[ERROR] url.Parse: " + err.Error())
			}

			if filenameURL == nil {
				continue
			}

			filename := pa.Base(filenameURL.Path)
			if filename == "." {
				continue
			}

			fullpath := filepath.Join("out", path, filename)
			if strings.HasPrefix(source, "/") {
				source = u + "/" + filename
				log.Println(fullpath)
			}
			log.Println(fullpath)
			if checkFileExists(fullpath) {
				continue
			}

			if err := downloadFile(fullpath, source); err != nil {
				log.Println("[ERROR] downloadFile: " + err.Error())
				continue
			}

			exists, err := findPackage(fullpath)
			if err != nil {
				log.Println("[ERROR] findPackage: " + err.Error())
				continue
			}

			log.Printf("exists?: %t \n", exists)
			if exists {
				log.Printf("log: %s \n", "url: "+u+"path: "+fullpath)
				if err := logToFile("url: " + u + "path: " + fullpath); err != nil {
					log.Println("[ERROR] logToFile: " + err.Error())
					return
				}
			}
		}

	}

	return

}

func (p *Package) readConfigFile() error {
	// ReadFile following statement is useful for reading small files,
	// 	don't use it for reading large files
	b, err := ioutil.ReadFile("package.yml")
	if err != nil {
		return err
	}
	p.ConfigFile = b

	if p.ConfigFile != nil {
		if err := yaml.Unmarshal(p.ConfigFile, p); err != nil {
			return err
		}
	}

	return nil
}

// logToFile creates a file and passes it log.SetOutput,
// and then logs the given message
func logToFile(message string) error {
	// create a log file
	f, err := os.OpenFile("package.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	log.SetOutput(f)
	log.Println(message)
	return nil
}

// findPackage finds a string ex:`"scripts": in given file.
func findPackage(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, err
		}
	}
	defer f.Close()

	const bufferSize = 100

	buffer := make([]byte, bufferSize)
	for {
		bytesRead, err := f.Read(buffer)
		if err != nil {
			if err != io.EOF {
				return false, err
			}
			break
		}
		if bytes.Contains(buffer[:bytesRead], []byte(`"scripts":`)) {
			return true, nil
		}
	}
	return false, nil
}

// checkFileExists checks if file already exists,
// helps avoiding downloading file, if it already exists
func checkFileExists(path string) bool {
	if _, err := os.Stat(path); err == nil || os.IsExist(err) {
		return true
	}
	return false
}

// downloadFile downloads the given url and returns it,
// to be used by other parts of program
func downloadFile(path, url string) error {
	// get the data
	res, err := http.Get(url)
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		return nil
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
	var outputDirectory string
	if domainName != "" {
		outputDirectory = domainName
	}
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
			log.Println("[+] New Header: " + values[0] + ": " + values[1])
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
		log.Printf("[!]%s returned an %d instead of %d\n", url, res.StatusCode, http.StatusOK)
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
