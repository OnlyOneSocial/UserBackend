package model

//Notification ...
type Notification struct {
	Type      string `json:"type"`
	UserID    int    `json:"userid"`
	UserName  string `json:"username"`
	TimeStamp int64  `json:"timestamp"`
}
