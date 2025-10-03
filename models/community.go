package models

import (
	"errors"
	"mychat/global"
)

type Community struct {
	Model
	Name   string
	SelfId uint
	Type   int
	Image  string
	Desc   string
}

func FindUsers(groupId uint) (*[]uint, error) {
	relation := make([]Relation, 0)
	if tx := global.DB.Where("other_id = ? and type = 2", groupId).Find(&relation); tx.RowsAffected == 0 {
		return nil, errors.New("未查询到成员信息")
	}
	selfIds := make([]uint, 0)
	for _, v := range relation {
		userId := v.SelfId
		selfIds = append(selfIds, userId)
	}

	return &selfIds, nil
}
