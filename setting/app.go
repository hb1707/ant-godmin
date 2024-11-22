package setting

import (
	"encoding/json"
	"log"
	"os"
)

type AppInfo struct {
	Name    string
	Version string
	Desc    string
	Channel string
}

func ReadAppInfo(path string) AppInfo {
	// Read app.json
	b, err := os.ReadFile(path)
	if err != nil {
		log.Println("[ERROR]", err)
		return AppInfo{}
	}
	// Unmarshal app.json
	var appInfo AppInfo
	err = json.Unmarshal(b, &appInfo)
	if err != nil {
		log.Println("[ERROR]", err)
		return AppInfo{}
	}
	return appInfo
}
