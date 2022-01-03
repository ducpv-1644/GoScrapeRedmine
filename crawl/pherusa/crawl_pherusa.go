package pherusa

import (
	"fmt"
	"github.com/gocolly/colly"
	"go-scrape-redmine/config"
	"go-scrape-redmine/crawl"
	"go-scrape-redmine/models"
	"gorm.io/gorm"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func NewPherusa(db *gorm.DB) crawl.Pherusa {
	return &Pherusa{
		db: db,
	}
}

type Pherusa struct {
	db *gorm.DB
}

func (a *Pherusa) CrawlIssuePherusa(projectId uint, version string) error {

	project := models.Project{}
	err := a.db.First(&project, projectId).Error
	if err != nil {
		return err
	}

	var cookies []*http.Cookie
	var cookie *http.Cookie
	var sessionId = [2]string{"62763d83d60df722e5bc9d462609d63d", "1ea15e2ff784549a4df1eea62fb62fab"}

	for _, e := range sessionId {
		cookie = &http.Cookie{
			Name:     os.Getenv("NAME_COOKIE"),
			Value:    e,
			Path:     "/",
			Domain:   os.Getenv("DOMAIN"),
			Secure:   true,
			HttpOnly: true,
		}
		cookies = append(cookies, cookie)
	}
	for i, _ := range cookies {
		var c = colly.NewCollector(
			colly.AllowedDomains(os.Getenv("DOMAIN")),
		)
		c.SetRequestTimeout(5000 * time.Second)
		fmt.Println("cookies", []*http.Cookie{cookies[i]})
		if err := c.SetCookies(os.Getenv("HOMEPAGE"), []*http.Cookie{cookies[i]}); err != nil {
			fmt.Println("Errors: have errors from cookies", err)
		}
		// code
		projectName := strings.ReplaceAll(project.Prefix, "/projects/", "")
		fmt.Println("projectName", projectName)
		CrawlIssue(c, a.db, projectName, version)
		//CrawlDueDateVersion(c, a.db, projectName)
		err = a.CreateVersion(version, projectId)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *Pherusa) CreateVersion(version string, projectId uint) error {
	versionProject := models.VersionProject{}
	err := a.db.Where("project_id = ? and version = ?", projectId, version).First(&versionProject).Error

	if err == gorm.ErrRecordNotFound {
		a.db.Create(&models.VersionProject{
			ProjectId: projectId,
			Version:   version,
			Current:   false,
		})
		return nil
	}
	if err != nil {
		return err
	}

	return nil
}

func initColly(url string) *colly.Collector {

	config.LoadENV()
	db := config.DBConnect()

	var cookies []*http.Cookie
	var cookie *http.Cookie
	//var sessionId = [2]string{"62763d83d60df722e5bc9d462609d63d", "1ea15e2ff784549a4df1eea62fb62fab"}
	sessionId := strings.Split(os.Getenv("VALUE_COOKIE"), ",")
	for _, e := range sessionId {
		cookie = &http.Cookie{
			Name:     os.Getenv("NAME_COOKIE"),
			Value:    e,
			Path:     "/",
			Domain:   os.Getenv("DOMAIN"),
			Secure:   true,
			HttpOnly: true,
		}
		cookies = append(cookies, cookie)
	}
	for i, _ := range cookies {
		var c = colly.NewCollector(
			colly.AllowedDomains(os.Getenv("DOMAIN")),
		)
		c.SetRequestTimeout(5000 * time.Second)
		fmt.Println("cookies", []*http.Cookie{cookies[i]})
		if err := c.SetCookies(url, []*http.Cookie{cookies[i]}); err != nil {
			fmt.Println("Errors: have errors from cookies", err)
		}
		CrawlProject(c, db)

	}

	return nil
}

func CrawlProject(c *colly.Collector, db *gorm.DB) {
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	c.OnHTML("a.project", func(e *colly.HTMLElement) {

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

	err := c.Visit(os.Getenv("HOMEPAGE") + "/projects")
	if err != nil {
		return
	}
	fmt.Println("Crwal project data finished.")
}

func CrawlIssue(c *colly.Collector, db *gorm.DB, project string, version string) {
	fmt.Println("Crawl issue")
	c.OnHTML("div.autoscroll tbody", func(e *colly.HTMLElement) {
		e.ForEach("tr", func(_ int, tr *colly.HTMLElement) {
			issue := models.Issue{
				IssueId:            tr.DOM.Children().Filter(".id").Text(),
				IssueProject:       tr.DOM.Children().Filter(".project").Text(),
				IssueTracker:       tr.DOM.Children().Filter(".tracker").Text(),
				IssueSubject:       tr.DOM.Children().Filter(".subject").Text(),
				IssueStatus:        tr.DOM.Children().Filter(".status").Text(),
				IssuePriority:      tr.DOM.Children().Filter(".priority").Text(),
				IssueAssignee:      tr.DOM.Children().Filter(".assigned_to").Text(),
				IssueTargetVersion: tr.DOM.Children().Filter(".fixed_version").Text(),
				IssueDueDate:       tr.DOM.Children().Filter(".due_date").Text(),
				IssueStartDate:     tr.DOM.Children().Filter(".start_date").Text(),
				IssueEstimatedTime: tr.DOM.Children().Filter(".estimated_hours").Text(),
				IssueDoneRatio:     tr.DOM.Children().Filter(".done_ratio").Text(),
				IssueSource:        "pherusa",
				IssueVersion:       version,
				IssueState:         "",
			}
			overDueId := tr.DOM.Filter(".overdue").Children().Filter(".id").Text()
			if overDueId != "" {
				issue.IssueState = "overdue"
			}

			var dbIssue models.Issue

			db.Find(&dbIssue, issue)

			if dbIssue == (models.Issue{}) {
				db.Create(&issue)
			}
			if dbIssue.IssueId == issue.IssueId {
				db.Model(&dbIssue).Where("issue_id = ?", issue.IssueId).Updates(issue)
			}

		})
	})

	err := c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		RandomDelay: 1 * time.Second,
	})
	if err != nil {
		return
	}

	fullURL := fmt.Sprintf(os.Getenv("HOMEPAGE") + "/projects/" + project + "/issues" + getUrlFromVersion(version))
	err = c.Visit(fullURL)
	fmt.Println("fullURL", fullURL)
	if err != nil {
		return
	}

	fmt.Println("Crwal issue data finished.")
}

func CrawlActivities(c *colly.Collector, db *gorm.DB) {
	year, month, day := time.Now().Date()
	dayStr := strconv.Itoa(day)
	monthStr := strconv.Itoa(int(month))
	yearStr := strconv.Itoa(year)
	if len(dayStr) == 1 {
		dayStr = "0" + dayStr
	}
	if len(monthStr) == 1 {
		monthStr = "0" + monthStr
	}
	dateNow := monthStr + "/" + dayStr + "/" + yearStr
	c.OnHTML("div#activity dl", func(e *colly.HTMLElement) {
		dateActivity := e.DOM.Prev().Text()
		if dateActivity == "Today" {
			dateActivity = dateNow
		}
		e.ForEach("dt", func(_ int, dt *colly.HTMLElement) {
			dd := dt.DOM.Next()

			activity := models.Activity{
				MemberId:    getMemberId(dt.ChildAttr(".gravatar", "src")),
				MemberName:  dd.Children().Filter("span.author").Text(),
				Project:     dt.ChildText("span.project"),
				Time:        dt.ChildText("span.time"),
				Date:        dateActivity,
				Issues:      dt.ChildText("a"),
				Type:        strings.TrimSpace(getTypeIssue(dt.Attr("class"))),
				Description: dd.Children().Filter("span.description").Text(),
			}

			var dbActivity models.Activity
			db.Find(&dbActivity, activity)
			if dbActivity == (models.Activity{}) {
				db.Create(&activity)
			}
		})
	})

	err := c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		RandomDelay: 1 * time.Second,
	})
	if err != nil {
		return
	}

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	err = c.Visit(os.Getenv("HOMEPAGE") + "/activity")
	if err != nil {
		return
	}
	fmt.Println("Crwal activity data finished.")
}

func CrawlDueDateVersion(c *colly.Collector, db *gorm.DB, project string) {

	fmt.Println("Crawl due date ")

	c.OnHTML("#tab-content-versions > table > tbody", func(e *colly.HTMLElement) {
		fmt.Println(c)
		fmt.Println(e)
		e.ForEach("tr", func(_ int, tr *colly.HTMLElement) {
			//issue := models.DueDateVersion{}
			fmt.Println(tr.DOM.Children().Filter(".date").Text())
			//db.Find(&dbIssue, issue)
			//
			//if dbIssue == (models.Issue{}) {
			//	db.Create(&issue)
			//}
			//if dbIssue.IssueId == issue.IssueId {
			//	db.Model(&dbIssue).Where("issue_id = ?", issue.IssueId).Updates(issue)
			//}

		})
	})

	err := c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		RandomDelay: 1 * time.Second,
	})
	if err != nil {
		return
	}

	fullURL := fmt.Sprintf(os.Getenv("HOMEPAGE") + "/projects/" + project + "/settings")
	err = c.Visit(fullURL)
	if err != nil {
		return
	}

	fmt.Println("Crwal due date data finished.")
}

func getMemberId(url string) string {
	if strings.Contains(url, "person.png") {
		return ""
	}
	re, _ := regexp.Compile(`(avatar\?id=)\d+`)
	result := strings.Split(string(re.Find([]byte(url))), "=")
	return result[1]
}

func getUrlFromVersion(version string) string {
	if version == "" {
		return ""
	} else {
		return "?set_filter=1&sort=id:desc&f[]=fixed_version_id&op[fixed_version_id]==&v[fixed_version_id][]=" + version + "&f[]=tracker_id&op[tracker_id]=!&v[tracker_id][]=4&v[tracker_id][]=15&f[]=&c[]=status&c[]=assigned_to&c[]=estimated_hours&c[]=spent_hours&c[]=start_date&c[]=due_date&c[]=fixed_version&group_by=&t[]="
	}
}

func getTypeIssue(t string) string {
	if strings.Contains(t, "icon-issue-closed") {
		return "Close Issue"
	}
	if strings.Contains(t, "icon-issue-note") {
		return "Note Issue"
	}
	if strings.Contains(t, "icon-issue-edit") {
		return "Edit Issue"
	}
	if strings.Contains(t, "icon-time-entry") {
		return "Estimate Issue"
	}
	if strings.Contains(t, "icon-issue ") {
		return "New Issue"
	}
	return t
}

func (a *Pherusa) CrawlPherusa() {
	fmt.Println("Cron running...crawling data.")
	initColly(os.Getenv("HOMEPAGE"))
	//CrawlProject(c, a.db)
	//CrawlActivities(c, a.db)
}
