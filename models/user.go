package models

import (
	"time"

	"gorm.io/gorm"
)

type Model struct {
	Id        uint `gorm:"primaryKey"`
	CreateAt  time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type User struct {
	Model
	Name          string
	Password      string
	Avatar        string
	Gender        string `gorm:"column:gender;default:male;type:varchar(6) comment 'male表示男, famale表示女'"`
	Phone         string `valid:"matches(^1[3-9]{1}\\d{9}$)"`
	Email         string `valid:"email"`
	Identity      string
	ClientIP      string `valid:"ipv4"`
	ClientPort    string
	Salt          string
	LoginTime     *time.Time `gorm:"column:login_time"`
	HeartBeatTime *time.Time `gorm:"columnt:heart_beat_time"`
	LoginOutTime  *time.Time `gorm:"column:login_out_time"`
	IsLoginOut    bool
	DeviceInfo    string
}

func (table *User) UserTableName() string {
	return "user"
}

type Relation struct {
	Model
	SelfId  uint // 当前用户id
	OtherId uint // 加入的群或者好友id
	Type    int  // 关系类型：1表示好友关系，2表示群关系
	Desc    string
}

func (r *Relation) RelTableName() string {
	return "relation"
}
