package Notify

type Message struct {
	MemberName string `json:"issue_assignee"`
	Report     Report `json:"report"`
}
type Report struct {
	Status string `json:"status"`
	Amount int64  `json:"amount"`
}
