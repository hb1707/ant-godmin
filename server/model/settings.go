package model

import (
	"github.com/hb1707/ant-godmin/pkg/log"
)

type Settings struct {
	Key   string `json:"key" gorm:"column:setting_key;type:varchar(100) not null;uniqueIndex:idx_key;default:'';comment:键"` //键
	Value string `json:"value" gorm:"column:setting_value;type:varchar(128) not null;"`
	TableBase
}

func NewSettings(where ...interface{}) *Settings {
	var t = new(Settings)
	if len(where) > 0 {
		t.DB = DB.Table("sys_users").Model(&Settings{}).Where(where[0], where[1:]...).Select("setting_key,setting_value")
	} else {
		t.DB = DB.Table("sys_users").Model(&Settings{}).Select("setting_key,setting_value")
	}
	return t
}
func (t *Settings) All(order string) []Settings {
	var list []Settings
	t.List(&list, order)
	return list
}
func (t *Settings) GetOne(order string) *Settings {
	var user Settings
	t.One(&user, order)
	return &user
}

func (t *Settings) Edit() *Settings {
	var user Settings
	t.Request(t)
	err := t.AddOrUpdate()
	if err != nil {
		log.Error(err)
	}
	return &user
}
