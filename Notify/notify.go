package Notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-scrape-redmine/config"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func NotiChatWork(version string) {

	receivers := []string{os.Getenv("MEMBER_ONE_NOTI_CHAT_WORK"), os.Getenv("MEMBER_TWO_NOTI_CHAT_WORK")}
	config.LoadENV()
	db := config.DBConnect()
	listReport := NewNotify(db).GetReportMember("pherusa", version)
	fmt.Println("message", strings.Join(listReport, "\n"))
	t1 := time.Now()
	timeStr := convertDateToString(&t1)

	bot := BotChatWork{
		Service:   os.Getenv("SERVICE_CHAT_WORK"),
		Channel:   os.Getenv("CHANNEL_CHAT_WORK"),
		Receivers: receivers,
		Message:   "[info][title]" + timeStr + ": [/title]" + strings.Join(listReport, "\n") + "[/info]",
	}
	body, _ := json.Marshal(bot)

	_, err := http.Post(os.Getenv("URL_NOTI"), "application/json", bytes.NewBuffer(body))
	if err != nil {
		//Failed to read response.

		panic(err)
	}
	return
}
func NotiSlack(version string) {

	receivers := []string{os.Getenv("MEMBER_ONE_NOTI_SLACK"), os.Getenv("MEMBER_TWO_NOTI_SLACK")}
	config.LoadENV()
	db := config.DBConnect()
	listReport := NewNotify(db).GetReportMember("pherusa", version)
	fmt.Println("message", strings.Join(listReport, "\n"))

	bot := BotChatWork{
		Service:   os.Getenv("SERVICE_SLACK_SLACK"),
		Channel:   os.Getenv("CHANNEL_SLACK"),
		Receivers: receivers,
		Message:   "Daily report: " + strings.Join(listReport, "\n") + "",
	}
	body, _ := json.Marshal(bot)

	_, err := http.Post(os.Getenv("URL_NOTI"), "application/json", bytes.NewBuffer(body))
	if err != nil {
		//Failed to read response.

		panic(err)
	}
	return
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
