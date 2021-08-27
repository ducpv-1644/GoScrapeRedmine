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
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct{}

type response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
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
	claims["role"] = role
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
		RespondWithJSON(w, resp.Code, resp)
		return
	}

	dbuser := newUserUsecase().FindUserByEmail(db, user.Email)
	if dbuser.Email != "" {
		resp.Code = http.StatusBadRequest
		resp.Message = "Email already in use!"
		RespondWithJSON(w, resp.Code, resp)
		return
	}

	user.Password, err = generatehashPassword(user.Password)
	if err != nil {
		resp.Code = http.StatusBadRequest
		resp.Message = "error in password hash!"
		RespondWithJSON(w, resp.Code, resp)
		return
	}

	userCreated := newUserUsecase().CreateUser(db, user)
	RespondWithJSON(w, http.StatusOK, userCreated)
}

func (a *UserHandler) SignIn(w http.ResponseWriter, r *http.Request) {
	resp := response{}
	db := config.DBConnect()

	var authdetails models.Authentication
	err := json.NewDecoder(r.Body).Decode(&authdetails)
	if err != nil {
		resp.Code = http.StatusBadRequest
		resp.Message = err.Error()
		RespondWithJSON(w, resp.Code, resp)
		return
	}

	authuser := newUserUsecase().FindUserByEmail(db, authdetails.Email)
	passwordOK := checkPasswordHash(authdetails.Password, authuser.Password)
	if !passwordOK || authuser.Email == "" {
		resp.Code = http.StatusBadRequest
		resp.Message = "Username or Password is incorrect"
		RespondWithJSON(w, resp.Code, resp)
		return
	}

	validToken, err := generateJWT(authuser.Email, authuser.Role)
	if err != nil {
		resp.Code = http.StatusBadRequest
		resp.Message = "Failed to generate token"
		RespondWithJSON(w, resp.Code, resp)
		return
	}

	var token models.Token
	token.Email = authuser.Email
	token.Role = authuser.Role
	token.TokenString = validToken
	RespondWithJSON(w, http.StatusOK, token)
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

type Effort struct {
	Projects string `json:"projects"`
	Date     string `json:"date"`
	Member   []Member
}

type Member struct {
	MemberName         string  `json:"membername"`
	TotalEstimatedTime float64 `json:"totalrstimatedtime"`
	TotalSpentTime     float64 `json:"totalspenttime"`
	Issue              models.Issue
}

func (a *UserHandler) GetEffort(w http.ResponseWriter, r *http.Request) {

	resp := response{}
	db := config.DBConnect()

	var effort Effort
	filter := r.URL.Query().Get("filter")
	projectName := r.URL.Query().Get("project")
	ranges := r.URL.Query().Get("range")
	splitRanges := strings.Split(ranges, "-")

	issue := []models.Issue{}

	switch filters := filter; {
	case filters == "effort":
		if len(splitRanges) == 1 {
			db.Where("issue_project = ?AND issue_start_date BETWEEN ? AND ?", projectName, splitRanges[0], splitRanges[0]).Find(&issue)
		} else {
			db.Where("issue_project = ? AND issue_start_date BETWEEN ? AND ?", projectName, splitRanges[0], splitRanges[1]).Find(&issue)
		}
	default:
		resp.Code = http.StatusBadRequest
		resp.Message = "filter invalid"
		RespondWithJSON(w, resp.Code, resp)
		return
	}

	var member []Member

	for _, elememt := range issue {

		issueClone := issue
		totalEstimatedTime := 0.0
		totalSpentTime := 0.0

		db.Where("issue_assignee = ? ", elememt.IssueAssignee).Find(&issueClone)
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
		}
		data := Member{
			MemberName:         elememt.IssueAssignee,
			TotalEstimatedTime: totalEstimatedTime,
			TotalSpentTime:     totalSpentTime,
			Issue:              elememt,
		}
		member = append(member, data)
	}

	effort = Effort{
		Projects: projectName,
		Date:     ranges,
		Member:   member,
	}

	RespondWithJSON(w, http.StatusOK, effort)
}
func (a *UserHandler) CrawData(w http.ResponseWriter, r *http.Request) {
	resp := response{}
	Redmine.NewRedmine().CrawlRedmine()
	resp.Code = http.StatusOK
	resp.Message = "Request crawl redmine data finished."
	RespondWithJSON(w, resp.Code, resp)
}
