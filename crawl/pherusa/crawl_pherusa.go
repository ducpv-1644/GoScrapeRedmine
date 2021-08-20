package pherusa

import (
	"fmt"
	"go-scrape-redmine/crawl"
)

func NewPherusa() crawl.Pherusa {
	return &Pherusa{}
}
type Pherusa struct{}

func (a *Pherusa) CrawlPherusa() {
	fmt.Println("Comming soon...")
}
