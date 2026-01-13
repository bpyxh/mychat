package models

type Message struct {
	Model
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
