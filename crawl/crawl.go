package crawl

import (
	"fmt"
	"go-scrape-redmine/models"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gocolly/colly"
	"gorm.io/gorm"
)

func InitColly(url string) *colly.Collector {
	c := colly.NewCollector(
		colly.AllowedDomains(os.Getenv("DOMAIN")),
	)
	c.SetRequestTimeout(400 * time.Second)
	var cookies []*http.Cookie
	cookie := &http.Cookie{
		Name:     os.Getenv("NAME_COOKIE"),
		Value:    os.Getenv("VALUE_COOKIE"),
		Path:     "/",
		Domain:   os.Getenv("DOMAIN"),
		Secure:   true,
		HttpOnly: true,
	}
	cookies = append(cookies, cookie)

	if err := c.SetCookies(url, cookies); err != nil {
		fmt.Println("Errors: have errors from cookies", err)

	}
	return c
}

func CrawlProject(wg *sync.WaitGroup, c *colly.Collector, db *gorm.DB) {

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	go c.OnHTML("a.project", func(e *colly.HTMLElement) {

		text := e.Text
		href := e.Attr("href")
		project := models.Project{
			Name:   text,
			Prefix: href,
		}
		var dbProejct models.Project
		db.Where("name = ?", project.Name).First(&dbProejct)
		if dbProejct.Name == "" {
			db.Create(&project)
		}
	})

	c.Visit(os.Getenv("HOMEPAGE") + "/projects")
	fmt.Println("crwal project done")
}
