package Notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-scrape-redmine/config"
	"net/http"
	"os"
)

type BotChatWork struct {
	Service   string   `json:"service"`
	Channel   string   `json:"channel"`
	Receivers []string `json:"receivers"`
	Message   string   `json:"message"`
}

func API() {

	receivers := []string{os.Getenv("MEMBER1"), os.Getenv("MEMBER2")}
	config.LoadENV()
	db := config.DBConnect()
	listReport := NewNotify(db).GetIssueOverdueStatusNone("pherusa")
	message, err := json.Marshal(listReport[0])
	if err != nil {
		fmt.Println("error during marshal message notify: ", err)
	}
	bot := BotChatWork{
		Service:   os.Getenv("CHATWORK"),
		Channel:   os.Getenv("CHANNEL"),
		Receivers: receivers,
		Message:   string(message),
	}
	//Nguyễn Văn A: New: 1 | InProgress: 2 | Overdue: 2 (1234, 5678) | NoEstimateTime: 1 (7890)
	body, _ := json.Marshal(bot)

	_, err = http.Post("http://10.0.4.171:5000/notify", "application/json", bytes.NewBuffer(body))
	if err != nil {
		//Failed to read response.
		panic(err)
	}
	return
}
