package main

import (
	"go-scrape-redmine/config"
	"go-scrape-redmine/crawl"
	"go-scrape-redmine/models"
	"go-scrape-redmine/server"
	"os"
	"sync"
)

const num_workers = 1

func main() {
	var wg sync.WaitGroup
	wg.Add(num_workers)
	config.LoadENV()
	db := config.DBConnect()
	models.DBMigrate(db)

	c := crawl.InitColly(os.Getenv("HOMEPAGE"))
	go crawl.CrawlProject(&wg, c, db)
	go server.Run(&wg)

	wg.Wait()
}
