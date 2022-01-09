package noti

import (
	"bytes"
	"encoding/json"
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
	message := "texx"
	bot := BotChatWork{
		Service:   os.Getenv("CHATWORK"),
		Channel:   os.Getenv("CHANNEL"),
		Receivers: receivers,
		Message:   message,
	}
	//Nguyễn Văn A: New: 1 | InProgress: 2 | Overdue: 2 (1234, 5678) | NoEstimateTime: 1 (7890)
	body, _ := json.Marshal(bot)

	_, err := http.Post("http://10.0.4.171:5000/notify", "application/json", bytes.NewBuffer(body))
	if err != nil {
		//Failed to read response.
		panic(err)
	}
	return
}
