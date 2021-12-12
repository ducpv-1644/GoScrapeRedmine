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
	} else if seed == "issue" {
		fmt.Println("Importing issue")
		//Redmine.NewRedmine().CrawlRedmine()
		pherusa.NewPherusa(db).CrawlPherusa()
		return
	} else if seed == "getissue" {
		Notify.NewNotify(db).GetIssueOverdueStatusNone("pherusa")
		return
	} else if seed == "apiIssue" {
		pherusa.NewPherusa(db).CrawlIssuePherusa(3, "854")
		return
	} else if seed == "noti" {
		Notify.NotiChatWork()
		Notify.NotiSlack()
		return
	} else if seed != "none" {
		fmt.Println("Flag seed invalid")
		return
	}

	var wg sync.WaitGroup
	wg.Add(num_workers)

	cr := cron.New()
	cr.AddFunc("0 18 * * *", Redmine.NewRedmine().CrawlRedmine)
	cr.AddFunc("0 18 * * *", Pherusa.NewPherusa(db).CrawlPherusa)
	cr.AddFunc("0 18 * * *", Notify.NotiChatWork)
	cr.AddFunc("0 18 * * *", Notify.NotiSlack)
	cr.Start()

	go server.Run(&wg)
	wg.Wait()
}
