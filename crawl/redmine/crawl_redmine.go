package redmine

import (
	"fmt"
	"go-scrape-redmine/config"
	"go-scrape-redmine/crawl"
	"go-scrape-redmine/models"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly"
	"gorm.io/gorm"
)

func NewRedmine() crawl.Redmine {
	return &Redmine{}
}

type Redmine struct{}

func getMemberId(url string) string {
	if strings.Contains(url, "person.png") {
		return ""
	}
	re, _ := regexp.Compile(`(avatar\?id=)\d+`)
	result := strings.Split(string(re.Find([]byte(url))), "=")
	return result[1]
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

func initColly(url string) *colly.Collector {
	c := colly.NewCollector(
		colly.AllowedDomains(os.Getenv("DOMAIN")),
	)
	c.SetRequestTimeout(400 * time.Second)
	var cookies []*http.Cookie
	cookie := &http.Cookie{
		Name:     os.Getenv("NAME_COOKIE"),
		Value:    os.Getenv("VALUE_COOKIE"),
		Path:     "/",
		Domain:   os.Getenv("DOMAIN"),
		Secure:   true,
		HttpOnly: true,
	}
	cookies = append(cookies, cookie)

	if err := c.SetCookies(url, cookies); err != nil {
		fmt.Println("Errors: have errors from cookies", err)

	}
	return c
}

var wg sync.WaitGroup

func crawlProject(c *colly.Collector, db *gorm.DB) {

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

	c.Visit(os.Getenv("HOMEPAGE") + "/projects")

	fmt.Println("Crwal project data finished.")
	wg.Done()
}

func crawlActivity(c *colly.Collector, db *gorm.DB) {

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

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		RandomDelay: 1 * time.Second,
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	c.Visit(os.Getenv("HOMEPAGE") + "/activity")
	fmt.Println("Crwal activity data finished.")
	wg.Done()
}

func crawlIssue(c *colly.Collector, db *gorm.DB) {
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
				IssueEstimatedTime: tr.DOM.Children().Filter(".estimated_hours").Text(),
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

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		RandomDelay: 1 * time.Second,
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	for i := 1; i <= 5; i++ {
		fullURL := fmt.Sprintf(os.Getenv("HOMEPAGE") + "/issues?page=" + strconv.Itoa(i))
		c.Visit(fullURL)
	}

	fmt.Println("Crwal issue data finished.")
	wg.Done()
}

func crawlIssueDetail(c *colly.Collector, db *gorm.DB) {
	var dbIssues models.Issue
	var issue_id []string
	var createdTime string
	var updatedTime string
	var update string
	var splitCreatedTime []string
	var splitUpdatedTime []string
	c.OnHTML("div#content", func(e *colly.HTMLElement) {

		desIssue := e.ChildAttrs("div.issue > p >a ", "title")

		if len(desIssue) == 1 {
			createdTime = desIssue[0]
			splitCreatedTime = strings.Split(createdTime, " ")
			fmt.Println(splitCreatedTime[0])

			updatedTime := [...]string{"null"}
			update = updatedTime[0]
			fmt.Println(update)

		} else {
			createdTime = desIssue[0]
			splitCreatedTime = strings.Split(createdTime, " ")
			fmt.Println(splitCreatedTime[0])
			updatedTime = desIssue[1]
			splitUpdatedTime = strings.Split(updatedTime, " ")
			update = splitUpdatedTime[0]
		}

		// id idIssues
		idIssues := e.DOM.Children().Filter("h2").Text()
		splitId := strings.Split(idIssues, "#")

		divAttributes := e.DOM.Children().Filter("div.issue").Children().Filter("div.attributes")
		splitContentLeft := divAttributes.Children().Filter("div.splitcontent").Children().Filter("div.splitcontentleft").Children()

		IssueSpentTime := splitContentLeft.Filter("div.spent-time.attribute").Children().Filter("div.value").Text()
		splitSpentTime := strings.Split(IssueSpentTime, " ")

		issue := models.Issue{
			IssueCategory:        splitContentLeft.Filter("div.category.attribute").Children().Filter("div.value").Text(),
			IssueActualEndDate:   splitContentLeft.Filter("div.cf_11.attribute").Children().Filter("div.value").Text(),
			IssueStoryPoint:      splitContentLeft.Filter("div.cf_7.attribute").Children().Filter("div.value").Text(),
			IssueLink:            "https://dev.sun-asterisk.com/issues/" + splitId[1],
			IssueActualStartDate: splitContentLeft.Filter("div.cf_10.attribute").Children().Filter("div.value").Text(),
			IssueGitUrl:          splitContentLeft.Filter("div.cf_23.attribute").Children().Filter("div.value").Text(),
			IssueQaDeadline:      splitContentLeft.Filter("div.cf_27.attribute").Children().Filter("div.value").Text(),
			IssueStartDate:       splitContentLeft.Filter("div.start-date.attribute").Children().Filter("div.value").Text(),
			IssueDoneRatio:       splitContentLeft.Filter("div.progress.attribute").Children().Filter("div.value").Children().Filter("p.percent").Text(),
			IssueSpentTime:       splitSpentTime[0],
			IssueCreated:         splitCreatedTime[0],
			IssueUpdated:         update,
			IssueSource:          "redmine",
		}

		db.Model(&dbIssues).Where("issue_id = ?", splitId[1]).Updates(issue)
	})

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		RandomDelay: 1 * time.Second,
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	db.Model(&dbIssues).Pluck("issue_id", &issue_id)
	for _, element := range issue_id {
		c.Visit(os.Getenv("HOMEPAGE") + "/issues/" + element)
	}

	fmt.Println("Crwal issue detail data finished.")

}

func (a *Redmine) CrawlRedmine() {
	fmt.Println("Cron running...crawling data.")
	db := config.DBConnect()
	c := initColly(os.Getenv("HOMEPAGE"))
	wg.Add(3)
	go crawlProject(c, db)
	go crawlActivity(c, db)
	go crawlIssue(c, db)
	wg.Wait()
	// fmt.Println("oke")
	crawlIssueDetail(c, db)

}
