package Notify

type Message struct {
	MemberName string         `json:"issue_assignee"`
	Report     map[string]int `json:"report"`
}
type BotChatWork struct {
	Service   string   `json:"service"`
	Channel   string   `json:"channel"`
	Receivers []string `json:"receivers"`
	Message   string   `json:"message"`
}
