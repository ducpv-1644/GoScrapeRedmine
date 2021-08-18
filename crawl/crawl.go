package crawl

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gocolly/colly"
)

func SaveFile(file string, data string) {
	fileName := file
	files, err := os.Create(fileName)
	if err != nil {
		log.Fatalf("Could not create %s", fileName)
	}
	defer files.Close()
	writer := csv.NewWriter(files)
	defer writer.Flush()
	writer.Write([]string{data})
}

func Crawl(wg *sync.WaitGroup) {

	url := "https://dev.sun-asterisk.com/projects/digmee-lab/issues?set_filter=1&tracker_id=4"
	url1 := "https://dev.sun-asterisk.com/projects"
	fileName := "redmine.csv"
	files, err := os.Create(fileName)
	if err != nil {
		log.Fatalf("Could not create %s", fileName)
	}
	defer files.Close()
	writer := csv.NewWriter(files)
	defer writer.Flush()
	c := colly.NewCollector(
		colly.AllowedDomains("dev.sun-asterisk.com"),
	)
	c.SetRequestTimeout(400 * time.Second)
	var cookies []*http.Cookie
	cookie := &http.Cookie{
		Name:     "_session_id",
		Value:    "null",
		Path:     "/",
		Domain:   "dev.sun-asterisk.com",
		Secure:   true,
		HttpOnly: true,
	}
	cookies = append(cookies, cookie)

	if err := c.SetCookies(url, cookies); err != nil {
		fmt.Println("Errors: have errors from cookies", err)
		return
	}

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	// lấy được project
	c.OnHTML("div.root", func(e *colly.HTMLElement) {
		a, _ := e.DOM.Html()
		fmt.Println(a)

	})
	c.Visit(url1)

}
