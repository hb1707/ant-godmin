package model

import (
	"github.com/hb1707/ant-godmin/consts"
)

type Logs struct {
	Uid     int                  `json:"uid" form:"uid" gorm:"column:uid;comment:用户ID;type:integer"`
	Action  consts.LogActionType `json:"action" form:"action" gorm:"column:action;comment:操作;type:varchar(10)"`
	TypeId  string               `json:"type_id" form:"type_id" gorm:"column:type_id;comment:类型;type:varchar(10)"`
	Title   string               `json:"title" form:"title" gorm:"column:title;comment:标题;type:text"`
	Content []byte               `json:"content" form:"content" gorm:"column:content;comment:详情;type:json"`
	TableBase
}

func NewLogs(where ...interface{}) *Logs {
	var t = new(Logs)
	if len(where) > 0 {
		t.DB = DB.Table("sys_logs").Model(&Logs{}).Where(where[0], where[1:]...)
	} else {
		t.DB = DB.Table("sys_logs").Model(&Logs{})
	}
	return t
}
