package handler

import (
	"encoding/json"
	"fmt"
	"go-scrape-redmine/app/users"
	userRepository "go-scrape-redmine/app/users/repository"
	userUsecase "go-scrape-redmine/app/users/usecase"
	"go-scrape-redmine/config"
	Redmine "go-scrape-redmine/crawl/redmine"
	"go-scrape-redmine/models"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct{}

type response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Result  interface{} `json:"result"`
}

func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func newUserUsecase() users.Usecase {
	return userUsecase.NewUserUsecase(
		userRepository.NewUserRepository(),
	)
}

func generatehashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func generateJWT(email, role string) (string, error) {
	var mySigningKey = []byte("unicorns")
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims["authorized"] = true
	claims["email"] = email
	claims["role"] = "user"
	claims["exp"] = time.Now().Add(time.Minute * 30).Unix()

	tokenString, err := token.SignedString(mySigningKey)

	if err != nil {
		fmt.Errorf("Something Went Wrong: %s", err.Error())
		return "", err
	}
	return tokenString, nil
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (a *UserHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	resp := response{}
	db := config.DBConnect()

	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		resp.Code = http.StatusBadRequest
		resp.Message = err.Error()
		RespondWithJSON(w, http.StatusOK, resp)
		return
	}

	if user.Email == "" {
		resp.Code = http.StatusBadRequest
		resp.Message = "Email is required!"
		RespondWithJSON(w, http.StatusOK, resp)
		return
	}

	dbuser := newUserUsecase().FindUserByEmail(db, user.Email)
	if dbuser.Email != "" {
		resp.Code = http.StatusBadRequest
		resp.Message = "Email already in use!"
		RespondWithJSON(w, http.StatusOK, resp)
		return
	}

	user.Password, err = generatehashPassword(user.Password)
	if err != nil {
		resp.Code = http.StatusBadRequest
		resp.Message = "error in password hash!"
		RespondWithJSON(w, http.StatusOK, resp)
		return
	}

	userCreated := newUserUsecase().CreateUser(db, user)
	resp.Code = http.StatusOK
	resp.Result = userCreated
	RespondWithJSON(w, http.StatusOK, resp)
}

func (a *UserHandler) SignIn(w http.ResponseWriter, r *http.Request) {

	resp := response{}
	db := config.DBConnect()

	var authdetails models.Authentication
	err := json.NewDecoder(r.Body).Decode(&authdetails)
	if err != nil {
		resp.Code = http.StatusBadRequest
		resp.Message = err.Error()
		RespondWithJSON(w, http.StatusOK, resp)
		return
	}

	authuser := newUserUsecase().FindUserByEmail(db, authdetails.Email)
	passwordOK := checkPasswordHash(authdetails.Password, authuser.Password)
	if !passwordOK || authuser.Email == "" {
		resp.Code = http.StatusBadRequest
		resp.Message = "Username or Password is incorrect"
		RespondWithJSON(w, http.StatusOK, resp)
		return
	}

	validToken, err := generateJWT(authuser.Email, authuser.Role)
	if err != nil {
		resp.Code = http.StatusBadRequest
		resp.Message = "Failed to generate token"
		RespondWithJSON(w, http.StatusOK, resp)
		return
	}

	var token models.Token
	token.Email = authuser.Email
	token.Role = authuser.Role
	token.TokenString = validToken

	resp.Code = http.StatusOK
	resp.Result = token
	RespondWithJSON(w, http.StatusOK, resp)
}
func weekStart(year, week int) time.Time {
	// Start from the middle of the year:
	t := time.Date(year, 7, 1, 0, 0, 0, 0, time.UTC)

	// Roll back to Monday:
	if wd := t.Weekday(); wd == time.Sunday {
		t = t.AddDate(0, 0, -6)
	} else {
		t = t.AddDate(0, 0, -int(wd)+1)
	}

	// Difference in weeks:
	_, w := t.ISOWeek()
	t = t.AddDate(0, 0, (week-w)*7)

	return t
}

func quarterRange(year, quarter int) (string, string) {

	var dayStart string
	var dayEnd string

	switch quarters := quarter; {
	case quarters == 1:
		yyyStart := strconv.Itoa(year)
		dayStart = "01/01/" + yyyStart
		yyyEnd := strconv.Itoa(year)
		dayEnd = "03/31/" + yyyEnd
	case quarters == 2:
		yyyStart := strconv.Itoa(year)
		dayStart = "04/01/" + yyyStart
		yyyEnd := strconv.Itoa(year)
		dayEnd = "06/30/" + yyyEnd
	case quarters == 3:
		yyyStart := strconv.Itoa(year)
		dayStart = "07/01/" + yyyStart
		yyyEnd := strconv.Itoa(year)
		dayEnd = "09/30/" + yyyEnd
	case quarters == 4:
		yyyStart := strconv.Itoa(year)
		dayStart = "10/01/" + yyyStart
		yyyEnd := strconv.Itoa(year)
		dayEnd = "12/31/" + yyyEnd
	}

	return dayStart, dayEnd

}
func weekRange(year, week int) (string, string) {

	start := weekStart(year, week)
	end := start.AddDate(0, 0, 6)
	ddStart := start.Day()
	mmStart := int(start.Month())
	yyyyStart := start.Year()

	var daystart string

	if ddStart < 10 {
		daystart = strconv.Itoa(mmStart) + "/" + "0" + strconv.Itoa(ddStart) + "/" + strconv.Itoa(yyyyStart)
	} else if mmStart < 10 {
		daystart = "0" + strconv.Itoa(mmStart) + "/" + strconv.Itoa(ddStart) + "/" + strconv.Itoa(yyyyStart)
	} else {
		daystart = strconv.Itoa(mmStart) + "/" + strconv.Itoa(ddStart) + "/" + strconv.Itoa(yyyyStart)
	}

	ddEnd := end.Day()
	mmEnd := int(end.Month())
	yyyyEnd := end.Year()

	var dayend string

	if ddEnd < 10 {
		dayend = strconv.Itoa(mmEnd) + "/" + "0" + strconv.Itoa(ddEnd) + "/" + strconv.Itoa(yyyyEnd)
	} else if mmEnd < 10 {
		dayend = "0" + strconv.Itoa(mmEnd) + "/" + strconv.Itoa(ddEnd) + "/" + strconv.Itoa(yyyyEnd)
	} else {
		dayend = strconv.Itoa(mmEnd) + "/" + strconv.Itoa(ddEnd) + "/" + strconv.Itoa(yyyyEnd)
	}

	return daystart, dayend
}
func (a *UserHandler) GetActivity(w http.ResponseWriter, r *http.Request) {
	resp := response{}
	db := config.DBConnect()

	memberID := r.URL.Query().Get("member")
	date := r.URL.Query().Get("date")
	filter := r.URL.Query().Get("filter")
	projectUrl := r.URL.Query().Get("project")
	week := r.URL.Query().Get("week")
	year := r.URL.Query().Get("year")
	weeks, _ := strconv.Atoi(week)
	years, _ := strconv.Atoi(year)
	quarter := r.URL.Query().Get("quarter")
	quarters, _ := strconv.Atoi(quarter)

	if memberID == "" || date == "" {
		resp.Code = http.StatusBadRequest
		resp.Message = "Member id or date is requied"
		RespondWithJSON(w, resp.Code, resp)
		return
	}

	var activity []models.Activity

	switch filters := filter; {
	case filters == "user":
		db.Where("member_id = ? ", memberID).Find(&activity)
	case filters == "date":
		db.Where("date = ? ", date).Find(&activity)
	case filters == "project":
		db.Where("project = ?", projectUrl).Find(&activity)
	case filters == "projectbymember":
		db.Where("member_id = ? AND project = ?", memberID, projectUrl).Find(&activity)
	case filters == "week":
		daystart, dayend := weekRange(years, weeks)
		fmt.Println(daystart)
		fmt.Println(dayend)
		db.Where("member_id = ? AND date BETWEEN ? AND ?", memberID, daystart, dayend).Find(&activity)
	case filters == "quarter":
		dayStart, dayEnd := quarterRange(years, quarters)
		db.Where("member_id = ? AND date BETWEEN ? AND ?", memberID, dayStart, dayEnd).Find(&activity)
	case filters == "both":
		db.Where("date = ? AND member_id = ?", date, memberID).Find(&activity)
	default:
		resp.Code = http.StatusBadRequest
		resp.Message = "filter invalid"
		RespondWithJSON(w, resp.Code, resp)
		return
	}
	RespondWithJSON(w, http.StatusOK, activity)

}

type EffortProject struct {
	Projects string           `json:"projects"`
	Date     string           `json:"date"`
	Issues   []IssueDataTable `json:"issues"`
	Members  []Member         `json:"members"`
}
type EffortMember struct {
	Member  string `json:"member"`
	Date    string `json:"date"`
	Porject []Porject
}
type Member struct {
	MemberName         string  `json:"membername"`
	TotalEstimatedTime float64 `json:"totalrstimatedtime"`
	TotalSpentTime     float64 `json:"totalspenttime"`
	TotalIssue         int     `json: "totalssue"`
	ListIssue          []ListIssue
}
type ListIssue struct {
	NameIssue string `json:"nameissue"`
	LinkIssue string `json:"linkissue"`
}
type Porject struct {
	ProjectName        string  `json:"projectname"`
	TotalEstimatedTime float64 `json:"totalrstimatedtime"`
	TotalSpentTime     float64 `json:"totalspenttime"`
	TotalIssue         int     `json: "totalssue"`
	ListIssue          []ListIssue
}
type ProjectName struct {
	Name   string `json:"name"`
	Prefix string `json:"prefix"`
}
type ProjectDetail struct {
	Name       string `json:"name"`
	TotalIssue int    `json: "totalssue"`
	ListIssue  []ListIssue
}
type IssueDataTable struct {
	IssueId            string `json:"issue_id"`
	IssueTracker       string `json:"issue_tracker"`
	IssueStatus        string `json:"issue_status"`
	IssuePriority      string `json:"issue_priority"`
	IssueSubject       string `json:"issue_subject"`
	IssueAssignee      string `json:"issue_assignee"`
	IssueTargetVersion string `json:"issue_target_version"`
	IssueDueDate       string `json:"issue_due_date"`
	IssueEstimatedTime string `json:"issue_estimated_time"`
	IssueDoneRatio     string `json:"issue_done_ratio"`
}

func removeDuplicateStr(strSlice []string) []string {
	allKeys := make(map[string]bool)
	list := []string{}
	for _, item := range strSlice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

func (a *UserHandler) GetEffort(w http.ResponseWriter, r *http.Request) {

	// dung chung
	resp := response{}
	db := config.DBConnect()
	var ranges string
	week := r.URL.Query().Get("week")
	year := r.URL.Query().Get("year")
	weeks, _ := strconv.Atoi(week)
	years, _ := strconv.Atoi(year)
	quarter := r.URL.Query().Get("quarter")
	quarters, _ := strconv.Atoi(quarter)
	filter := r.URL.Query().Get("filter")
	ranges = r.URL.Query().Get("range")
	splitRanges := strings.Split(ranges, "-")
	projectID := r.URL.Query().Get("project_id")
	memberName := r.URL.Query().Get("member")
	issue := []models.Issue{}
	project := models.Project{}

	// khao bao struct
	var effortProject EffortProject
	var effortMember EffortMember
	var dayStartWeek string
	var dayEndWeek string
	var dayStartQuarter string
	var dayEndQuarter string
	var projectName string
	// xu li tim kiem theo project
	db.Where("id = ?", projectID).Find(&project)
	projectName = project.Name
	if projectName != "" {

		var filterInValid bool
		var listNameMember []string
		filterInValid = false

		switch filters := filter; {
		case filters == "week":
			dayStartWeek, dayEndWeek = weekRange(years, weeks)
			ranges = dayStartWeek + "-" + dayEndWeek
		case filters == "quarter":
			dayStartQuarter, dayEndQuarter = quarterRange(years, quarters)
			ranges = dayStartQuarter + "-" + dayEndQuarter
		case filters == "effort":
			filterInValid = false
		default:
			filterInValid = true
		}
		if filterInValid {
			resp.Code = http.StatusBadRequest
			resp.Message = "'filter' invalid"
			RespondWithJSON(w, http.StatusOK, resp)
			return
		}

		db.Where("issue_project = ?", projectName).Find(&issue)

		issueDatatable := []IssueDataTable{}
		members := []Member{}

		// lay ra all name member
		for _, elememt := range issue {
			listNameMember = append(listNameMember, elememt.IssueAssignee)
		}

		// member nao da co ten thi xu li trung
		dataNameMember := removeDuplicateStr(listNameMember)

		// xu li tong time member lam va name+link Issue
		for _, elementName := range dataNameMember {
			var listIssue []ListIssue

			issueClone := issue
			totalEstimatedTime := 0.0
			totalSpentTime := 0.0
			if filter == "week" {
				db.Where("issue_assignee = ? AND issue_due_date BETWEEN ? AND ?", elementName, dayStartWeek, dayEndWeek).Find(&issueClone)
			}
			if filter == "quarter" {
				db.Where("issue_assignee = ? AND issue_due_date BETWEEN ? AND ?", elementName, dayStartQuarter, dayEndQuarter).Find(&issueClone)
			}
			if filter == "effort" {
				if len(splitRanges) == 1 {
					db.Where("issue_assignee = ? AND issue_due_date = ?", elementName, splitRanges[0]).Find(&issueClone)
				} else {
					db.Where("issue_assignee = ? AND issue_due_date BETWEEN ? AND ?", elementName, splitRanges[0], splitRanges[1]).Find(&issueClone)
				}
			}
			for _, elementIssue := range issueClone {

				estTime := 0.0
				spentime := 0.0
				if elementIssue.IssueEstimatedTime != "" {
					estTime, _ = strconv.ParseFloat(elementIssue.IssueEstimatedTime, 64)
				}
				if elementIssue.IssueSpentTime != "" {
					spentime, _ = strconv.ParseFloat(elementIssue.IssueSpentTime, 64)
				}
				totalEstimatedTime = totalEstimatedTime + estTime
				totalSpentTime = totalSpentTime + spentime

				dataListIssue := ListIssue{
					NameIssue: elementIssue.IssueSubject,
					LinkIssue: elementIssue.IssueLink,
				}
				listIssue = append(listIssue, dataListIssue)
				issueDatatable = append(issueDatatable, IssueDataTable{
					IssueId:            elementIssue.IssueId,
					IssueTracker:       elementIssue.IssueTracker,
					IssueStatus:        elementIssue.IssueStatus,
					IssuePriority:      elementIssue.IssuePriority,
					IssueSubject:       elementIssue.IssueSubject,
					IssueAssignee:      elementIssue.IssueAssignee,
					IssueTargetVersion: elementIssue.IssueTargetVersion,
					IssueDueDate:       elementIssue.IssueDueDate,
					IssueEstimatedTime: elementIssue.IssueEstimatedTime,
					IssueDoneRatio:     elementIssue.IssueDoneRatio,
				})
			}
			data := Member{
				MemberName:         elementName,
				TotalEstimatedTime: totalEstimatedTime,
				TotalSpentTime:     totalSpentTime,
				TotalIssue:         len(listIssue),
				ListIssue:          listIssue,
			}
			members = append(members, data)
		}

		effortProject = EffortProject{
			Projects: projectName,
			Date:     ranges,
			Issues:   issueDatatable,
			Members:  members,
		}
		resp.Code = http.StatusOK
		resp.Result = effortProject
		RespondWithJSON(w, http.StatusOK, resp)
		return
	}

	// xu li tim kiem theo member
	if memberName != "" {
		switch filters := filter; {
		case filters == "effort":
			db.Where("issue_assignee = ?", memberName).Find(&issue)
		case filters == "week":
			{
				dayStartWeek, dayEndWeek = weekRange(years, weeks)
				ranges = dayStartWeek + "-" + dayEndWeek
				db.Where("issue_assignee = ?", memberName).Find(&issue)
			}
		case filters == "quarter":
			{
				dayStartQuarter, dayEndQuarter = quarterRange(years, quarters)
				ranges = dayStartQuarter + "-" + dayEndQuarter
				db.Where("issue_assignee = ?", memberName).Find(&issue)
			}

		default:
			resp.Code = http.StatusBadRequest
			resp.Message = "filter invalid"
			RespondWithJSON(w, resp.Code, resp)
			return
		}

		var project []Porject
		var listNameProject []string

		for _, elememt := range issue {
			listNameProject = append(listNameProject, elememt.IssueAssignee)
		}

		dataNameProject := removeDuplicateStr(listNameProject)

		for _, e := range dataNameProject {
			var listIssue []ListIssue
			var nameProject string
			issueClone := issue
			totalEstimatedTime := 0.0
			totalSpentTime := 0.0
			if filter == "week" {
				db.Where("issue_assignee = ? AND issue_due_date BETWEEN ? AND ?", e, dayStartWeek, dayEndWeek).Find(&issueClone)
			}
			if filter == "quarter" {
				db.Where("issue_assignee = ? AND issue_due_date BETWEEN ? AND ?", e, dayStartQuarter, dayEndQuarter).Find(&issueClone)
			}
			if filter == "effort" {
				if len(splitRanges) == 1 {
					db.Where("issue_assignee = ? AND issue_due_date = ?", e, splitRanges[0]).Find(&issueClone)
				} else {
					db.Where("issue_assignee = ? AND issue_due_date BETWEEN ? AND ?", e, splitRanges[0], splitRanges[1]).Find(&issueClone)
				}
			}
			for _, elementIssue := range issueClone {
				estTime := 0.0
				spentime := 0.0
				if elementIssue.IssueEstimatedTime != "" {
					estTime, _ = strconv.ParseFloat(elementIssue.IssueEstimatedTime, 64)
				}
				if elementIssue.IssueSpentTime != "" {
					spentime, _ = strconv.ParseFloat(elementIssue.IssueSpentTime, 64)
				}
				totalEstimatedTime = totalEstimatedTime + estTime
				totalSpentTime = totalSpentTime + spentime
				nameProject = elementIssue.IssueProject
				dataListIssue := ListIssue{
					NameIssue: elementIssue.IssueSubject,
					LinkIssue: elementIssue.IssueLink,
				}
				listIssue = append(listIssue, dataListIssue)
			}

			data := Porject{
				ProjectName:        nameProject,
				TotalEstimatedTime: totalEstimatedTime,
				TotalSpentTime:     totalSpentTime,
				TotalIssue:         len(listIssue),
				ListIssue:          listIssue,
			}
			project = append(project, data)
		}

		effortMember = EffortMember{
			Member:  memberName,
			Date:    ranges,
			Porject: project,
		}
		RespondWithJSON(w, http.StatusOK, effortMember)
	} else {
		// tim kiem all project
		var listdataDetail []ListIssue
		var projetDetail []ProjectDetail
		projectIssue := []models.Project{}
		issueClone := issue
		project := r.URL.Query().Get("projects")
		pre := "/projects/" + project

		db.Where("prefix = ?", pre).Find(&projectIssue)
		for _, e := range projectIssue {
			data := ProjectName{
				Name:   e.Name,
				Prefix: e.Prefix,
			}
			nameProject := data.Name

			db.Where("issue_project = ?", nameProject).Find(&issueClone)
			for _, elementIssueDetail := range issueClone {
				dataListIssue := ListIssue{
					NameIssue: elementIssueDetail.IssueSubject,
					LinkIssue: elementIssueDetail.IssueLink,
				}
				listdataDetail = append(listdataDetail, dataListIssue)
			}

			dataDetail := ProjectDetail{
				Name:       nameProject,
				TotalIssue: len(listdataDetail),
				ListIssue:  listdataDetail,
			}
			projetDetail = append(projetDetail, dataDetail)
		}

		RespondWithJSON(w, http.StatusOK, projetDetail)
	}
}

func (a *UserHandler) CrawData(w http.ResponseWriter, r *http.Request) {
	resp := response{}
	Redmine.NewRedmine().CrawlRedmine()
	resp.Code = http.StatusOK
	resp.Message = "Request crawl redmine data finished."
	RespondWithJSON(w, resp.Code, resp)
}

func (a *UserHandler) GetAllProject(w http.ResponseWriter, r *http.Request) {
	db := config.DBConnect()
	dbprojects := []models.Project{}
	db.Find(&dbprojects)

	resp := response{}
	resp.Code = http.StatusOK
	resp.Result = dbprojects
	RespondWithJSON(w, http.StatusOK, resp)
}

type GetIssueByMember struct {
	SumSpentTime     float64 `json:"sum_spent_time"`
	SumEstimatedTime float64 `json:"sum_est_time"`
	IssueResult      []getAllIssueResult
}
type getAllIssueResult struct {
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
	IssueDoneRatio     string `json:"issue_done_ratio"`
}

type MemberIssue struct {
	MemberId    string `json:"menberid"`
	MemberName  string `json:"menbername"`
	MemberEmail string `json:"menberemail"`
}

func (a *UserHandler) GetAllIssue(w http.ResponseWriter, r *http.Request) {
	ranges := r.URL.Query().Get("ranges")
	splitRanges := strings.Split(ranges, "-")

	db := config.DBConnect()
	issue := []models.Issue{}
	member := models.Member{}
	var result []getAllIssueResult
	var issueclone = issue
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		fmt.Println("id is missing in parameters")
	}
	db.Where("member_id = ?", id).First(&member)
	db.Where("issue_assignee=? AND issue_due_date BETWEEN ? AND ?", member.MemberName, splitRanges[0], splitRanges[1]).Find(&issue)

	sumEstimated := 0.0
	sumSpent := 0.0
	for _, issue := range issue {
		estTime := 0.0
		spentTime := 0.0
		if issue.IssueEstimatedTime != "" {
			estTime, _ = strconv.ParseFloat(issue.IssueEstimatedTime, 64)
		}
		if issue.IssueSpentTime != "" {
			spentTime, _ = strconv.ParseFloat(issue.IssueEstimatedTime, 64)
		}
		sumEstimated = sumEstimated + estTime
		sumSpent = sumSpent + spentTime

	}
	db.Where("issue_assignee=? AND issue_due_date BETWEEN ? AND ?", member.MemberName, splitRanges[0], splitRanges[1]).Find(&issueclone)
	for _, e := range issueclone {
		issueData := getAllIssueResult{
			IssueId:            e.IssueId,
			IssueProject:       e.IssueProject,
			IssueTracker:       e.IssueTracker,
			IssueSubject:       e.IssueSubject,
			IssueStatus:        e.IssueStatus,
			IssuePriority:      e.IssuePriority,
			IssueAssignee:      e.IssueAssignee,
			IssueTargetVersion: e.IssueTargetVersion,
			IssueDueDate:       e.IssueDueDate,
			IssueEstimatedTime: e.IssueEstimatedTime,
			IssueDoneRatio:     e.IssueDoneRatio,
		}
		result = append(result, issueData)
	}
	listIssueByMember := GetIssueByMember{
		SumSpentTime:     sumSpent,
		SumEstimatedTime: sumEstimated,
		IssueResult:      result,
	}
	resp := response{}
	resp.Code = http.StatusOK
	resp.Result = listIssueByMember
	RespondWithJSON(w, http.StatusOK, resp)
}

type getAllMemberResult struct {
	MemberID         string
	MemberName       string
	ProjectName      string
	SumSpentTime     float64
	SumEstimatedTime float64
}

func (a *UserHandler) GetAllMember(w http.ResponseWriter, r *http.Request) {
	ranges := r.URL.Query().Get("ranges")
	splitRanges := strings.Split(ranges, "-")

	db := config.DBConnect()
	var result []getAllMemberResult
	dbissues := []models.Issue{}
	dbmembers := []models.Member{}

	db.Find(&dbmembers)

	for _, member := range dbmembers {
		var dbnameProjects []string
		dbissuesClone := dbissues

		dbQuery := db.Model(&dbissuesClone).Where("issue_assignee = ?", member.MemberName)
		dbQuery.Pluck("issue_project", &dbnameProjects)
		nameProjectUniq := removeDuplicateStr(dbnameProjects)

		issuesWithMember := []models.Issue{}
		db.Where("issue_assignee = ? AND issue_due_date BETWEEN ? AND ?", member.MemberName, splitRanges[0], splitRanges[1]).Find(&issuesWithMember)
		sumEstimated := 0.0
		sumSpent := 0.0
		for _, issue := range issuesWithMember {
			estTime := 0.0
			spentTime := 0.0
			if issue.IssueEstimatedTime != "" {
				estTime, _ = strconv.ParseFloat(issue.IssueEstimatedTime, 64)
			}
			if issue.IssueSpentTime != "" {
				spentTime, _ = strconv.ParseFloat(issue.IssueEstimatedTime, 64)
			}
			sumEstimated = sumEstimated + estTime
			sumSpent = sumSpent + spentTime
		}

		memberData := getAllMemberResult{
			MemberID:         member.MemberId,
			MemberName:       member.MemberName,
			ProjectName:      strings.Join(nameProjectUniq, ", "),
			SumSpentTime:     sumSpent,
			SumEstimatedTime: sumEstimated,
		}
		result = append(result, memberData)
	}

	resp := response{}
	resp.Code = http.StatusOK
	resp.Result = result
	RespondWithJSON(w, http.StatusOK, resp)
}
