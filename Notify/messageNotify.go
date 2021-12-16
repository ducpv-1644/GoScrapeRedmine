package Notify

type Message struct {
	MemberName string         `json:"issue_assignee"`
	Report     map[string]int `json:"report"`
}
type BotChatWork struct {
	Service     string      `json:"service"`
	Channel     string      `json:"channel"`
	Receivers   []string    `json:"receivers"`
	Message     string      `json:"message"`
	Attachments []Attachments `json:"attachments"`
}

type Attachments struct {
	Color  string `json:"color"`
	Blocks []Block  `json:"blocks"`
}
type MessageBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type Block struct {
	Type string       `json:"type"`
	Text MessageBlock `json:"text"`
}
