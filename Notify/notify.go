package Notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-scrape-redmine/config"
	"net/http"
	"os"
	"strings"
)

func NotiChatWork() {

	receivers := []string{os.Getenv("MEMBER_ONE_NOTI_CHAT_WORK"), os.Getenv("MEMBER_TWO_NOTI_CHAT_WORK")}
	config.LoadENV()
	db := config.DBConnect()
	listReport := NewNotify(db).GetIssueOverdueStatusNone("pherusa", "854")
	fmt.Println("message", strings.Join(listReport, "\n"))

	bot := BotChatWork{
		Service:   os.Getenv("SERVICE_CHAT_WORK"),
		Channel:   os.Getenv("CHANNEL_CHAT_WORK"),
		Receivers: receivers,
		Message:   "[info][title] Daily report: [/title]" + strings.Join(listReport, "\n") + "[/info]",
	}
	body, _ := json.Marshal(bot)

	_, err := http.Post(os.Getenv("URL_NOTI"), "application/json", bytes.NewBuffer(body))
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
	listReport := NewNotify(db).GetIssueOverdueStatusNone("pherusa", "854")
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
