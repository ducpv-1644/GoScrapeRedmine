package crawl

type Redmine interface {
	CrawlRedmine()
}

type Pherusa interface {
	CrawlPherusa()
	CrawlIssuePherusa(projectId uint, version string) error
}
