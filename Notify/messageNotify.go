package Notify

type Message struct {
	MemberName string         `json:"issue_assignee"`
	Report     map[string]int `json:"report"`
}
