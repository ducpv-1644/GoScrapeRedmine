package Notify

import (
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

func (n notify) CreateConfig(config models.ConfigNoty) (models.ConfigNoty, error) {
	//TODO implement me
	err := n.db.Create(&config).Error
	if err != nil {
		return models.ConfigNoty{}, err
	}
	return config, nil
}

func (n notify) UpdateConfig(config models.ConfigNoty) (models.ConfigNoty, error) {
	//TODO implement me
	old := models.ConfigNoty{}
	err := n.db.Where("id = ?", config.ID).First(&old).Error
	if err != nil {
		return models.ConfigNoty{}, err
	}
	err = n.db.Save(config).Error
	if err != nil {
		return models.ConfigNoty{}, err
	}

	return config, nil
}

func (n notify) GetAllConfig(projectId string) ([]models.ConfigNoty, error) {
	//TODO implement me
	results := make([]models.ConfigNoty, 0)
	err := n.db.Where("project_id = ?", projectId).Find(&results).Error
	if err != nil {
		return []models.ConfigNoty{}, err
	}

	return results, nil
}

func (n notify) GetConfigById(id string) (models.ConfigNoty, error) {
	//TODO implement me
	result := models.ConfigNoty{}
	err := n.db.Where("id = ?", id).First(&result).Error
	if err != nil {
		return models.ConfigNoty{}, err
	}

	return result, nil
}

func (n notify) DeleteConfig(id string) error {
	//TODO implement me
	config := models.ConfigNoty{}
	err := n.db.Where("id = ?", id).First(&config).Error
	if err != nil {
		return err
	}
	err = n.db.Delete(&config).Error
	if err != nil {
		return err
	}

	return nil
}

func (n notify) GetReportMember(source string, version string) ([]Block, string) {
	//TODO implement me
	listIssue := make([]models.Issue, 0)
	blocks := make([]Block, 0)
	fmt.Println("version: ", version)
	err := n.db.Where("issue_source = ? AND issue_version = ?  AND issue_tracker != 'EPIC' and issue_tracker != 'story'", source, version).Find(&listIssue).Error
	if err != nil {
		fmt.Println("error during get issue: ", err)
		return blocks, ""
	}

	listMemberMap := make(map[string]string, 0)
	listMember := make([]string, 0)

	//status := []string{NoEstimate, NoSpentTime, NoDueDate, OverDue, Doing, Free}
	targetVersion := ""

	for _, issue := range listIssue {
		targetVersion = issue.IssueTargetVersion
		if issue.IssueAssignee != "" && issue.IssueAssignee != listMemberMap[issue.IssueAssignee] {
			listMemberMap[issue.IssueAssignee] = issue.IssueAssignee
		}

	}

	for _, member := range listMemberMap {
		listMember = append(listMember, member)
	}

	for _, member := range listMember {

		str := make([]string, 0)
		overDueArr := make([]string, 0)
		noEstimateArr := make([]string, 0)
		noSpentTImeArr := make([]string, 0)
		noDueTimeArr := make([]string, 0)
		freeArr := make([]string, 0)
		noFreeArr := make([]string, 0)
		for _, issue := range listIssue {
			if issue.IssueAssignee == member {
				if issue.IssueStartDate != "" && issue.IssueEstimatedTime == "" {
					startDate, err := convertStringToTime(issue.IssueStartDate)
					if err != nil {
						fmt.Println("error during convert string to time: ", err)
					}
					if startDate.After(time.Now()) {
						noEstimateArr = append(noEstimateArr, issue.IssueId)
					}
				}

				if issue.IssueDueDate != "" && issue.IssueSpentTime == "" {
					dueDate, err := convertStringToTime(issue.IssueDueDate)
					if err != nil {
						fmt.Println("error during convert string to time: ", err)

					}
					if dueDate.After(time.Now()) || dueDate.Equal(time.Now()) {
						noSpentTImeArr = append(noSpentTImeArr, issue.IssueId)
					}
				}

				if issue.IssueStartDate != "" && issue.IssueDueDate == "" {
					startDate, err := convertStringToTime(issue.IssueStartDate)
					if err != nil {
						fmt.Println("error during convert string to time: ", err)

					}
					if startDate.After(time.Now()) || startDate.Equal(time.Now()) {
						noDueTimeArr = append(noDueTimeArr, issue.IssueId)
					}
				}

				if checkFree(issue.IssueStatus) {
					freeArr = append(freeArr, issue.IssueId)
				}

				if !checkFree(issue.IssueStatus) {
					noFreeArr = append(noFreeArr, issue.IssueId)
				}

				if issue.IssueDueDate != "" {

					if issue.IssueState == "overdue" {
						overDueArr = append(overDueArr, issue.IssueId)
					}
				}

			}

		}
		if len(noEstimateArr) > 0 {
			noEstimate := make([]string, 0)
			for i := 0; i < len(noEstimateArr); i++ {
				noEstimate = append(noEstimate, GetLinkIssue(noEstimateArr[i]))
			}
			noEstimatestr := "(" + strings.Join(noEstimate, ",") + ")"
			str = append(str, "*"+NoEstimate+":"+strconv.Itoa(len(noEstimateArr))+noEstimatestr+"*")
		}
		if len(noSpentTImeArr) > 0 {
			noSpentTIme := make([]string, 0)
			for i := 0; i < len(noSpentTImeArr); i++ {
				noSpentTIme = append(noSpentTIme, GetLinkIssue(noSpentTImeArr[i]))
			}
			noSpentTImestr := "(" + strings.Join(noSpentTIme, ",") + ")"
			str = append(str, "*"+NoSpentTime+": "+strconv.Itoa(len(noSpentTImeArr))+noSpentTImestr+"*")
		}
		if len(noDueTimeArr) > 0 {
			noDue := make([]string, 0)
			for i := 0; i < len(overDueArr); i++ {
				noDue = append(noDue, GetLinkIssue(overDueArr[i]))
			}
			noDuestr := "(" + strings.Join(noDue, ",") + ")"
			str = append(str, "*"+NoDueDate+": "+strconv.Itoa(len(noDueTimeArr))+noDuestr+"*")
		}
		if len(overDueArr) > 0 {
			dueDate := make([]string, 0)
			for i := 0; i < len(overDueArr); i++ {
				dueDate = append(dueDate, GetLinkIssue(overDueArr[i]))
			}
			dueDatestr := "(" + strings.Join(dueDate, ",") + ")"
			str = append(str, "*"+OverDue+": "+strconv.Itoa(len(overDueArr))+dueDatestr+"*")
		}
		if len(noFreeArr) > 0 && len(freeArr) > 0 {
			str = append(str, Doing+": "+strconv.Itoa(len(noFreeArr)))
		}
		if len(freeArr) > 0 && len(noFreeArr) == 0 {

			str = append(str, Free)
		}

		if len(str) > 0 {
			blocks = append(blocks, Block{
				Type: "section",
				Text: MessageBlock{
					Type: "mrkdwn",
					Text: "- " + member + ": " + strings.Join(str, " | "),
				},
			})
		}
	}

	return blocks, targetVersion
}

func (n notify) CreateSession(session models.SessionId) (models.SessionId, error) {
	//TODO implement me
	err := n.db.Create(&session).Error
	if err != nil {
		return models.SessionId{}, err
	}
	return session, nil
}

func (n notify) GetAllSession(sessionId string) ([]models.SessionId, error) {
	//TODO implement me
	results := make([]models.SessionId, 0)
	err := n.db.Where("id = ?", sessionId).Find(&results).Error
	if err != nil {
		return []models.SessionId{}, err
	}
	return results, nil
}

func (n notify) GetSessionById(id string) (models.SessionId, error) {
	//TODO implement me
	result := models.SessionId{}
	err := n.db.Where("id = ?", id).First(&result).Error
	if err != nil {
		return models.SessionId{}, err
	}

	return result, nil
}

type Notify interface {
	GetReportMember(source string, version string) ([]Block, string)
	CreateConfig(config models.ConfigNoty) (models.ConfigNoty, error)
	UpdateConfig(config models.ConfigNoty) (models.ConfigNoty, error)
	GetAllConfig(projectId string) ([]models.ConfigNoty, error)
	GetConfigById(id string) (models.ConfigNoty, error)
	DeleteConfig(id string) error
	CreateSession(config models.SessionId) (models.SessionId, error)
	GetAllSession(projectId string) ([]models.SessionId, error)
	GetSessionById(id string) (models.SessionId, error)
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

func NewNotify(db *gorm.DB) Notify {
	return notify{db: db}
}

func GetLinkIssue(id string) string {
	if id == "" {
		return ""
	}

	return "<https://pherusa-redmine.sun-asterisk.vn/issues/" + id + "|" + id + ">"
}
