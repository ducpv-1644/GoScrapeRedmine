package Notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-scrape-redmine/config"
	"go-scrape-redmine/models"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

//func NotiChatWork() {
//
//	receivers := []string{os.Getenv("MEMBER_ONE_NOTI_CHAT_WORK"), os.Getenv("MEMBER_TWO_NOTI_CHAT_WORK")}
//	config.LoadENV()
//	db := config.DBConnect()
//
//	version, err := GetCurrentVersion(db)
//	if err != nil {
//		fmt.Println("error during noti slack: ", err)
//		return
//	}
//
//	_, targetVersion := NewNotify(db).GetReportMember("pherusa", version)
//	//fmt.Println("message", strings.Join(listReport, "\n"))
//	t1 := time.Now()
//	timeStr := convertDateToString(&t1)
//
//	bot := BotChatWork{
//		Service:   os.Getenv("SERVICE_CHAT_WORK"),
//		Channel:   os.Getenv("CHANNEL_CHAT_WORK"),
//		Receivers: receivers,
//		Message:   "[info][title]" + targetVersion + "-" + timeStr + ": [/title]" + "[/info]",
//	}
//	body, _ := json.Marshal(bot)
//
//	_, err = http.Post(os.Getenv("URL_NOTI"), "application/json", bytes.NewBuffer(body))
//	if err != nil {
//		//Failed to read response.
//
//		panic(err)
//	}
//	return
//}
func NotiSlack(db *gorm.DB, receiversStr, projectId, service, channelId string) {

	receivers := strings.Split(receiversStr, ",")
	version, err := GetCurrentVersion(db, projectId)
	if err != nil {
		fmt.Println("error during noti slack: ", err)
		return
	}
	listReport, targetVersion := NewNotify(db).GetReportMember("pherusa", version)
	t1 := time.Now()
	timeStr := convertDateToString(&t1)
	var attachments = make([]Attachments, 0)
	attachments = append(attachments, Attachments{
		Color:  "#f2c744",
		Blocks: listReport,
	})
	bot := BotChatWork{
		Service:     service,
		Channel:     channelId,
		Receivers:   receivers,
		Message:     "*" + targetVersion + "-" + timeStr + ":" + "*" + "\n",
		Attachments: attachments,
	}
	body, _ := json.Marshal(bot)
	fmt.Println("body",bytes.NewBuffer(body))
	_, err = http.Post(os.Getenv("URL_NOTI"), "application/json", bytes.NewBuffer(body))
	if err != nil {
		//Failed to read response.
		panic(err)
	}
	return
}

func GetCurrentVersion(db *gorm.DB, projectId string) (string, error) {
	version := models.VersionProject{}
	err := db.Where("current = true and project_id = ?", projectId).Find(&version).Error
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

func NotyReports() {
	config.LoadENV()
	db := config.DBConnect()
	notyService := NewNotify(db)
	projectIds := strings.Split(os.Getenv("NOTI_PROJECT_IDS"), ",")
	//fmt.Println("projectIds: ", strings.Split(projectIds, ","))
	fmt.Println("projectIds", projectIds)
	configs := make([]models.ConfigNoty, 0)

	for _, id := range projectIds {
		configArr, err := notyService.GetAllConfig(id)
		if err != nil {
			fmt.Println("error during get project_id:", err)
		}
		for _, noty := range configArr {
			configs = append(configs, noty)
		}
	}
	//NotiSlack(db, "B121505", "3", "slack", "C02H99CP3H9")
	for _, configNoty := range configs {
		//fmt.Println("configNoty", configNoty.MemberId, configNoty.ProjectId, configNoty.Service, configNoty.ChannelId)
		NotiSlack(db, configNoty.MemberId, configNoty.ProjectId, configNoty.Service, configNoty.ChannelId)
	}
}
