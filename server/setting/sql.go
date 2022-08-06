package setting

import (
	"github.com/hb1707/ant-godmin/model"
	"time"
)

var Sql = map[string]string{}
var cacheTime = time.Now()

func reLoad() {
	list := model.NewSettings().All("id desc")
	for _, settings := range list {
		Sql[settings.Key] = settings.Value
	}
	cacheTime = time.Now()
}

func Get(k string) string {
	if _, exist := Sql[k]; exist || cacheTime.Add(time.Minute*10).Before(time.Now()) {
		reLoad()
	}
	return Sql[k]
}
func Set(k string, v string) {
	model.NewSettings("setting_key = ?", k).UpdateFieldNotId("setting_value", v)
	Sql[k] = v
}
