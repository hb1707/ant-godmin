package model

import (
	"time"

	"github.com/hb1707/ant-godmin/pkg/log"
)

type Settings struct {
	Key   string `json:"key" gorm:"column:setting_key;type:varchar(100) not null;uniqueIndex:idx_key;default:'';comment:键"` //键
	Value string `json:"value" gorm:"column:setting_value;type:varchar(256) not null;"`
	TableBase
}

func NewSettings(where ...interface{}) *Settings {
	var t = new(Settings)
	if len(where) > 0 {
		t.DB = DB.Model(&Settings{}).Where(where[0], where[1:]...)
	} else {
		t.DB = DB.Model(&Settings{})
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

func (t *Settings) Edit(must ...string) *Settings {
	var user Settings
	t.Request(t)
	err := t.AddOrUpdate(must)
	if err != nil {
		log.Error(err)
	}
	return &user
}

var SettingsCache = map[string]Settings{}
var cacheTime = time.Now()

func reLoad() {
	list := NewSettings().All("id desc")
	for _, settings := range list {
		SettingsCache[settings.Key] = settings
	}
	cacheTime = time.Now()
}

func SettingGet(k string, timeOut time.Duration) string {
	if _, exist := SettingsCache[k]; !exist || cacheTime.Add(time.Minute*10).Before(time.Now()) {
		reLoad()
	}
	if _, exist := SettingsCache[k]; !exist {
		sql := NewSettings()
		var up Settings
		up.Key = k
		up.Value = ""
		sql.Request(&up)
		sql.AddOrUpdate([]string{"setting_key", "setting_value"})
		up.UpdatedAt = time.Now()
		SettingsCache[k] = up
	}
	if timeOut == 0 || SettingsCache[k].UpdatedAt.Add(timeOut).After(time.Now()) {
		return SettingsCache[k].Value
	} else {
		return ""
	}
}

func SettingSet(k string, v string) {
	NewSettings("setting_key = ?", k).UpdateFieldNotId("setting_value", v)
	SettingsCache[k] = Settings{
		Key:   k,
		Value: v,
	}
}
