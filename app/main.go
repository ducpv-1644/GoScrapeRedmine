package main

import (
	"go-scrape-redmine/config"
	"go-scrape-redmine/models"
	"go-scrape-redmine/server"
	"sync"


)

const num_workers = 1

func main() {
	var wg sync.WaitGroup
	wg.Add(num_workers)

	config.LoadENV()
	db := config.DBConnect()
	models.DBMigrate(db)
	go server.Run(&wg)

	wg.Wait()
}
