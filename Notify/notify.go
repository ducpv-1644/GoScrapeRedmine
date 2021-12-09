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
	listReport := NewNotify(db).GetIssueOverdueStatusNone("pherusa")
	fmt.Println("message", strings.Join(listReport, "\n"))

	bot := BotChatWork{
		Service:   os.Getenv("SERVICE_CHAT_WORK"),
		Channel:   os.Getenv("CHANNEL_CHAT_WORK"),
		Receivers: receivers,
		Message:   "[info]" + strings.Join(listReport, "\n") + "[info]",
	}
	body, _ := json.Marshal(bot)

	_, err := http.Post(os.Getenv("URL_NOTI"), "application/json", bytes.NewBuffer(body))
	if err != nil {
		//Failed to read response.

		panic(err)
	}
	return
}
