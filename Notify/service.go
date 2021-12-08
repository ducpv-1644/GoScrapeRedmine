package Notify

import (
	"fmt"
	"go-scrape-redmine/models"
	"gorm.io/gorm"
	"strconv"
	"strings"
	"time"
)

type notify struct {
	db *gorm.DB
}

func (n notify) GetIssueOverdueStatusNone(source string) {
	//TODO implement me
	listIssue := make([]models.Issue, 0)
	err := n.db.Where("issue_source = ?", source).Find(&listIssue).Error
	if err != nil {
		fmt.Println("error during get issue: ", err)
		return
	}
	//listMember := make([]string, 0)
	listStatus := make([]string, 0)
	for _, issue := range listIssue {
		//if len(listMember) == 0 {
		//	listMember = append(listMember, issue.IssueAssignee)
		//} else {
		//	for _, member := range listMember {
		//		if issue.IssueAssignee != "" && issue.IssueAssignee != member {
		//			listMember = append(listMember, issue.IssueAssignee)
		//
		//		}
		//	}
		//}
		if len(listStatus) == 0 && issue.IssueStatus != "" {
			listStatus = append(listStatus, issue.IssueStatus)
		} else {
			for _, status := range listStatus {
				if issue.IssueStatus != "" && issue.IssueStatus != status {
					listStatus = append(listStatus, issue.IssueStatus)

				}
			}
		}

		//if issue.IssueStatus == "" {
		//	countNoStatus++
		//}
		//if issue.IssueEstimatedTime == "" {
		//	countNoEstimate++
		//}
		//if issue.IssueStatus == "New" && issue.IssueDueDate != "" {
		//	issueDate, err := convertStringToTime(issue.IssueDueDate)
		//	if err == nil && issueDate.After(time.Now()) {
		//		countOverDue++
		//	}
		//	if err != nil {
		//		fmt.Println("error during check overdue: ", err)
		//	}
		//}s
	}

	//for _, member := range listMember {
	//	fmt.Println(member)
	//}
	for _, member := range listStatus {
		fmt.Println(member)
	}

}

type Notify interface {
	GetIssueOverdueStatusNone(source string)
}

func convertStringToTime(date string) (time.Time, error) {
	split := strings.Split(date, "/")
	year, err := strconv.Atoi(split[2])
	if err != nil {
		return time.Time{}, err
	}
	month, err := strconv.Atoi(split[0])
	if err != nil {
		return time.Time{}, err
	}
	day, err := strconv.Atoi(split[1])
	if err != nil {
		return time.Time{}, err
	}

	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC), nil

}

func NewNotify(db *gorm.DB) Notify {
	return notify{db: db}
}
