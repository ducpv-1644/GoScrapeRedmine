package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name     string `json:"name"`
	Email    string `gorm:"unique" json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type Authentication struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Token struct {
	Role        string `json:"role"`
	Email       string `json:"email"`
	TokenString string `json:"token"`
}

type Project struct {
	gorm.Model
	Name   string  `json:"name"`
	Prefix string  `json:"prefix`
	Issue  []Issue `gorm:"many2many:project_issue;"`
}

type Member struct {
	gorm.Model
	MemberId    string `json:"menberid"`
	MemberName  string `json:"menbername"`
	MemberEmail string `json:"menberemail"`
}

type Activity struct {
	gorm.Model
	MemberId    string `json:"menberid"`
	MemberName  string `json:"menbername"`
	Project     string `json:"project"`
	Time        string `json:"time"`
	Date        string `json:"date"`
	Issues      string `json:"issues"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

type Issue struct {
	gorm.Model
	IssueId            string `json:"issue_id"`
	IssueProject       string `json:"issue_project"`
	IssueTracker       string `json:"issue_tracker"`
	IssueSubject       string `json:"issue_subject"`
	IssueStatus        string `json:"issue_status"`
	IssuePriority      string `json:"issue_priority"`
	IssueAssignee      string `json:"issue_assignee"`
	IssueTargetVersion string `json:"issue_target_version"`
	IssueDueDate       string `json:"issue_due_date"`
	IssueEstimatedTime string `json:"issue_estimated_time"`

	IssueCategory        string `json:"issue_category"`
	IssueStoryPoint      string `json:"issue_story_point"`
	IssueLink            string `json:"issue_link"`
	IssueActualStartDate string `json:"issue_actual_start_date"`
	IssueActualEndDate   string `json:"issue_actual_end_date"`
	IssueGitUrl          string `json:"issue_git_url"`
	IssueQaDeadline      string `json:"issueq_a_deadline"`
	IssueStartDate       string `json:"issue_start_date"`
	IssueDoneRatio       string `json:"issue_done_ratio"`
	IssueSpentTime       string `json:"issue_spent_time"`
	IssueAuthor          string `json:"issue_author"`
	IssueCreated         string `json:"issue_created"`
	IssueUpdated         string `json:"issue_updated"`
	IssueSource          string `json:"issue_source"`
	IssueVersion         string `json:"issue_version"`
	IssueTarget          string `json:"issue_target"`
}

type VersionProject struct {
	gorm.Model
	IdProject uint   `json:"id_project"`
	Version   string `json:"version"`
}

func DBMigrate(db *gorm.DB) {
	db.AutoMigrate(&User{})
	db.AutoMigrate(&Project{})
	db.AutoMigrate(&Member{})
	db.AutoMigrate(&Activity{})
	db.AutoMigrate(&Issue{})
	db.AutoMigrate(&VersionProject{})
}
