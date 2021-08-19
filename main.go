package main

import (
	"go-scrape-redmine/config"
	"go-scrape-redmine/crawl"
	"go-scrape-redmine/models"
	"go-scrape-redmine/server"
	"sync"

	"github.com/robfig/cron/v3"
)

const num_workers = 1

func main() {
	var wg sync.WaitGroup
	wg.Add(num_workers)

	config.LoadENV()
	db := config.DBConnect()
	models.DBMigrate(db)

	cr := cron.New()
	cr.AddFunc("0 18 * * *", crawl.CrawlData)
	cr.Start()

	go server.Run(&wg)

	wg.Wait()
}
