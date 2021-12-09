package Notify

import (
	"encoding/json"
	"fmt"
	"go-scrape-redmine/models"
	"gorm.io/gorm"
	"strconv"
	"strings"
	"time"
)

var (
	statusNew          = "New"
	statusInProgress   = "In Progress"
	statusInvestigated = "Investigated & Estimated"
	statusResolved     = "Resolved"
	statusPending      = "Pending"
	statusOverdue      = "Overdue"
)

type notify struct {
	db *gorm.DB
}

func (n notify) GetIssueOverdueStatusNone(source string) []string {
	//TODO implement me
	listIssue := make([]models.Issue, 0)
	sArray := make([]string, 0)
	err := n.db.Where("issue_source = ?", source).Find(&listIssue).Error
	if err != nil {
		fmt.Println("error during get issue: ", err)
		return sArray
	}
	listMemberMap := make(map[string]string, 0)
	listMember := make([]string, 0)

	status := []string{statusNew, statusResolved, statusPending, statusInvestigated, statusInvestigated}

	for _, issue := range listIssue {
		if issue.IssueAssignee != "" && issue.IssueAssignee != listMemberMap[issue.IssueAssignee] {
			listMemberMap[issue.IssueAssignee] = issue.IssueAssignee
		}

	}

	for _, member := range listMemberMap {
		listMember = append(listMember, member)
	}

	for _, member := range listMember {
		result := Message{
			MemberName: member,
			Report:     nil,
		}

		reports := make(map[string]int, 0)
		for i := 0; i < len(status); i++ {
			reports[status[i]] = 0
		}

		for _, issue := range listIssue {
			if issue.IssueAssignee == result.MemberName {
				switch issue.IssueStatus {
				case statusNew:
					if issue.IssueDueDate != "" {
						issueDate, err := convertStringToTime(issue.IssueDueDate)
						if err == nil && issueDate.Before(time.Now()) {
							reports["Overdue"]++
						} else {
							reports[issue.IssueStatus]++
						}
						if err != nil {
							fmt.Println("error during check overdue: ", err)
						}
					} else {
						reports[issue.IssueStatus]++
					}
					break
				case statusInvestigated:
					reports[issue.IssueStatus]++
					break
				case statusPending:
					reports[issue.IssueStatus]++
					break
				case statusInProgress:
					reports[issue.IssueStatus]++
					break
				case statusResolved:
					reports[issue.IssueStatus]++
					break
				default:
					break
				}
			}

		}
		result.Report = reports
		message, err := json.Marshal(reports)
		if err != nil {
			fmt.Println("error during marshal")
		}
		str := "-" + result.MemberName + ":" + string(message)
		str = strings.ReplaceAll(str, "\"", "")
		str = strings.ReplaceAll(str, "{", " ")
		str = strings.ReplaceAll(str, "}", " ")
		str = strings.ReplaceAll(str, ",", " | ")
		str = strings.ReplaceAll(str, "\\u0026", "&")
		sArray = append(sArray, str)
	}
	for _, s := range sArray {
		fmt.Println(s)
	}
	return sArray
}

type Notify interface {
	GetIssueOverdueStatusNone(source string) []string
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
