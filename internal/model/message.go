package model

import "gorm.io/gorm"

type Message struct {
	gorm.Model
	ID      uint64 `gorm:"column:id"`
	FormId  int64  `gorm:"column:from_id"`
	ToId    int64  `gorm:"column:to_id"`
	Type    int
	Media   int
	Content string
	Pic     string `json:"url"`
	Url     string
	Desc    string
	Amount  int
}

func (m *Message) MsgTableName() string {
	return "message"
}
