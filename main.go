package main

import (
    "flag"
    "fmt"
    "github.com/robfig/cron/v3"
    "go-scrape-redmine/Notify"
    "go-scrape-redmine/config"
    "go-scrape-redmine/crawl/pherusa"
    Pherusa "go-scrape-redmine/crawl/pherusa"
    "go-scrape-redmine/models"
    "go-scrape-redmine/server"
)


func main() {

	config.LoadENV()
	db := config.DBConnect()
	models.DBMigrate(db)
	var seed string
	flag.StringVar(&seed, "seed", "none", "seed option")
	flag.Parse()

	//if seed == "member" {
	//	fmt.Println("Importing member")
	//	Member.NewMember().SeedMember()
	//	return
	//} else
	if seed == "issue" {
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
	    Notify.NewNotiReport(db).NotyReports()
		return
	} else if seed != "none" {
		fmt.Println("Flag seed invalid")
		return
	}


	cr := cron.New()
	//_, err := cr.AddFunc("0 15 * * *", Redmine.NewRedmine().CrawlRedmine)
	//if err != nil {
	//	return
	//}
	_, err := cr.AddFunc("0 15 * * *", Pherusa.NewPherusa(db).CrawlPherusa)
	if err != nil {
		return
	}
	_, err = cr.AddFunc("0 16 * * *", Notify.NewNotiReport(db).NotyReports)
	if err != nil {
		return
	}
	cr.Start()

	 server.Run(db)
}
