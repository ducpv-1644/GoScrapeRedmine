package main

import (
	"flag"
	"fmt"
	"go-scrape-redmine/config"
	Redmine "go-scrape-redmine/crawl/redmine"
	"go-scrape-redmine/models"
	Member "go-scrape-redmine/seed/members"
	"go-scrape-redmine/server"
	"sync"

	"github.com/robfig/cron/v3"
)

const num_workers = 1

func main() {
	config.LoadENV()
	db := config.DBConnect()
	models.DBMigrate(db)

	var seed string
	flag.StringVar(&seed, "seed", "none", "seed option")
	flag.Parse()

	if seed == "member" {
		fmt.Println("Importing member")
		Member.NewMember().SeedMember()
		return
	} else if seed != "none" {
		fmt.Println("Flag seed invalid")
		return
	}

	var wg sync.WaitGroup
	wg.Add(num_workers)

	cr := cron.New()
	cr.AddFunc("0 18 * * *", Redmine.NewRedmine().CrawlRedmine)
	cr.Start()

	go server.Run(&wg)
	wg.Wait()
}
