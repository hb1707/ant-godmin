package setting

import (
	"encoding/json"
	"github.com/hb1707/ant-godmin/pkg/log"
	"os"
)

type AppInfo struct {
	Name    string
	Version string
	Desc    string
}

func ReadAppInfo(path string) AppInfo {
	// Read app.json
	b, err := os.ReadFile(path)
	if err != nil {
		log.Error(err)
		return AppInfo{}
	}
	// Unmarshal app.json
	var appInfo AppInfo
	err = json.Unmarshal(b, &appInfo)
	if err != nil {
		log.Error(err)
		return AppInfo{}
	}
	return appInfo
}
