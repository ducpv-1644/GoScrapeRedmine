package main

import (
	"flag"
	"fmt"
	"go-scrape-redmine/Notify"
	"go-scrape-redmine/config"
	"go-scrape-redmine/crawl/pherusa"
	Pherusa "go-scrape-redmine/crawl/pherusa"
	Redmine "go-scrape-redmine/crawl/redmine"
	"go-scrape-redmine/models"
	Member "go-scrape-redmine/seed/members"
	"go-scrape-redmine/server"
	"sync"

	"github.com/robfig/cron/v3"
)

const numWorkers = 1

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
	} else if seed == "issue" {
		fmt.Println("Importing issue")
		//Redmine.NewRedmine().CrawlRedmine()
		pherusa.NewPherusa(db).CrawlPherusa()
		return
	} else if seed == "getIssue" {
		Notify.NewNotify(db).GetReportMember("pherusa", "854")
		return
	} else if seed == "apiIssue" {
		err := pherusa.NewPherusa(db).CrawlIssuePherusa(3, "854")
		if err != nil {
		    fmt.Println("err",err)
			return
		}
		return
	} else if seed == "noti" {
		Notify.NotyReports()
		return
	} else if seed != "none" {
		fmt.Println("Flag seed invalid")
		return
	}

	var wg sync.WaitGroup
	wg.Add(numWorkers)

	cr := cron.New()
	_, err := cr.AddFunc("0 15 * * *", Redmine.NewRedmine().CrawlRedmine)
	if err != nil {
		return
	}
	_, err = cr.AddFunc("0 15 * * *", Pherusa.NewPherusa(db).CrawlPherusa)
	if err != nil {
		return
	}
	_, err = cr.AddFunc("0 16 * * *", Notify.NotyReports)
	if err != nil {
		return
	}
	cr.Start()

	go server.Run(&wg)
	wg.Wait()
}
