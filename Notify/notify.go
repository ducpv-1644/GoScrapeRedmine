package Notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-scrape-redmine/config"
	"go-scrape-redmine/models"
	"gorm.io/gorm"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func NotiChatWork() {

	receivers := []string{os.Getenv("MEMBER_ONE_NOTI_CHAT_WORK"), os.Getenv("MEMBER_TWO_NOTI_CHAT_WORK")}
	config.LoadENV()
	db := config.DBConnect()

	version, err := GetCurrentVersion(db)
	if err != nil {
		fmt.Println("error during noti slack: ", err)
		return
	}

	listReport, targetVersion := NewNotify(db).GetReportMember("pherusa", version)
	fmt.Println("message", strings.Join(listReport, "\n"))
	t1 := time.Now()
	timeStr := convertDateToString(&t1)

	bot := BotChatWork{
		Service:   os.Getenv("SERVICE_CHAT_WORK"),
		Channel:   os.Getenv("CHANNEL_CHAT_WORK"),
		Receivers: receivers,
		Message:   "[info][title]" + targetVersion + "-" + timeStr + ": [/title]" + strings.Join(listReport, "\n") + "[/info]",
	}
	body, _ := json.Marshal(bot)

	_, err = http.Post(os.Getenv("URL_NOTI"), "application/json", bytes.NewBuffer(body))
	if err != nil {
		//Failed to read response.

		panic(err)
	}
	return
}
func NotiSlack() {

	receivers := []string{os.Getenv("MEMBER_ONE_NOTI_SLACK"), os.Getenv("MEMBER_TWO_NOTI_SLACK")}
	config.LoadENV()
	db := config.DBConnect()
	version, err := GetCurrentVersion(db)
	if err != nil {
		fmt.Println("error during noti slack: ", err)
		return
	}
	listReport, targetVersion := NewNotify(db).GetReportMember("pherusa", version)
	fmt.Println("message", strings.Join(listReport, "\n"))
	t1 := time.Now()
	timeStr := convertDateToString(&t1)

	bot := BotChatWork{
		Service:   os.Getenv("SERVICE_SLACK_SLACK"),
		Channel:   os.Getenv("CHANNEL_SLACK"),
		Receivers: receivers,
		//Message:   "Daily report: " + strings.Join(listReport, "\n") + "",
		Message: targetVersion + "-" + timeStr + ":" + strings.Join(listReport, "\n") + "[/info]",
	}
	body, _ := json.Marshal(bot)

	_, err = http.Post(os.Getenv("URL_NOTI"), "application/json", bytes.NewBuffer(body))
	if err != nil {
		//Failed to read response.

		panic(err)
	}
	return
}

func GetCurrentVersion(db *gorm.DB) (string, error) {
	version := models.VersionProject{}
	err := db.Where("current = true").First(&version).Error
	if err != nil {
		return "", err
	}

	return version.Version, nil
}

func convertDateToString(time *time.Time) string {
	if time == nil {
		return ""
	}

	date := strconv.Itoa(time.Day())
	month := strconv.Itoa(int(time.Month()))
	year := strconv.Itoa(time.Year())
	return date + "/" + month + "/" + year
}
