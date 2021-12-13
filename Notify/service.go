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

const (
	NoEstimate  = "NoEstimate"
	NoSpentTime = "NoSpentTime"
	NoDueDate   = "NoDueDate"
	OverDue     = "OverDue"
	Doing       = "Doing"
	Free        = "Free"

	Closed   = "Closed"
	Resolved = "Resolved"
	Pending  = "Pending"
	Rejected = "Rejected"
)

type notify struct {
	db *gorm.DB
}

func (n notify) GetIssueOverdueStatusNone(source string, version string) []string {
	//TODO implement me
	listIssue := make([]models.Issue, 0)
	sArray := make([]string, 0)
	err := n.db.Where("issue_source = ? AND issue_tracker != 'EPIC' and issue_tracker != 'story'", source, version).Find(&listIssue).Error
	if err != nil {
		fmt.Println("error during get issue: ", err)
		return sArray
	}
	listMemberMap := make(map[string]string, 0)
	listMember := make([]string, 0)

	status := []string{NoEstimate, NoSpentTime, NoDueDate, OverDue, Doing, Free}

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

		overDue := make([]string, 0)

		for _, issue := range listIssue {
			if issue.IssueAssignee == result.MemberName {

				if issue.IssueStartDate != "" && issue.IssueEstimatedTime == "" {
					startDate, err := convertStringToTime(issue.IssueStartDate)
					if err != nil {
						fmt.Println("error during convert string to time: ", err)

					}
					if startDate.After(time.Now()) {
						reports[NoEstimate]++
					}
				}

				if issue.IssueDueDate != "" && issue.IssueSpentTime == "" {
					dueDate, err := convertStringToTime(issue.IssueDueDate)
					if err != nil {
						fmt.Println("error during convert string to time: ", err)

					}
					if dueDate.After(time.Now()) {
						reports[NoSpentTime]++
					}
				}

				if issue.IssueStartDate != "" && issue.IssueDueDate == "" {
					startDate, err := convertStringToTime(issue.IssueStartDate)
					if err != nil {
						fmt.Println("error during convert string to time: ", err)

					}
					if startDate.After(time.Now()) {
						reports[NoDueDate]++
					}
				}

				if checkFree(issue.IssueStatus) {
					reports[Free]++
				}

				if !checkFree(issue.IssueStatus) {
					reports[Doing]++
				}

				if issue.IssueDueDate != "" {
					dueDate, err := convertStringToTime(issue.IssueDueDate)
					if err != nil {
						fmt.Println("error during convert string to time: ", err)

					}
					if dueDate.Before(time.Now()) && !checkFree(issue.IssueStatus) {
						overDue = append(overDue, issue.IssueId)
						reports[OverDue]++
					}
				}

			}

		}
		messageReport := make(map[string]int, 0)
		for i := 0; i < len(status); i++ {
			if reports[status[i]] > 0 {
				messageReport[status[i]]++
			}
		}
		message, err := json.Marshal(messageReport)
		if err != nil {
			fmt.Println("error during marshal")
		}

		strMessage := string(message)

		if len(overDue) > 0 {
			strMessage = strMessage + "(" + strings.Join(overDue, " ") + ")"
		}

		str := formatData(result.MemberName, strMessage)
		sArray = append(sArray, str)
	}
	for _, s := range sArray {
		fmt.Println(s)
	}
	return sArray
}

type Notify interface {
	GetIssueOverdueStatusNone(source string, version string) []string
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

func checkFree(status string) bool {
	if status == Closed || status == Resolved || status == Pending || status == Rejected {
		return true
	} else {
		return false
	}
}

func formatData(memberName, message string) string {
	str := "-" + memberName + ":" + message
	str = strings.ReplaceAll(str, "\"", "")
	str = strings.ReplaceAll(str, "{", " ")
	str = strings.ReplaceAll(str, "}", " ")
	str = strings.ReplaceAll(str, ",", " | ")
	str = strings.ReplaceAll(str, "\\u0026", "&")
	return str
}

func NewNotify(db *gorm.DB) Notify {
	return notify{db: db}
}
