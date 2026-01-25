package model

import "gorm.io/gorm"

type Message struct {
	gorm.Model
	FormId   int64 `json:"fromId"`
	TargetId int64 `json:"targetId"`
	Type     int
	Media    int
	Content  string
	Pic      string `json:"url"`
	Url      string
	Desc     string
	Amount   int
}

func (m *Message) MsgTableName() string {
	return "message"
}
