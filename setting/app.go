package setting

import (
	"encoding/json"
	"log"
	"os"
)

type AppInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Desc    string `json:"desc"`
	Channel string `json:"channel"`
}

// ReadAppInfo 读取app.json
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

// WriteAppInfo 写入app.json
func WriteAppInfo(path string, appInfo AppInfo) {
	var exist = ReadAppInfo(path)
	if appInfo.Name != "" {
		exist.Name = appInfo.Name
	}
	if appInfo.Version != "" {
		exist.Version = appInfo.Version
	}
	if appInfo.Desc != "" {
		exist.Desc = appInfo.Desc
	}
	if appInfo.Channel != "" {
		exist.Channel = appInfo.Channel
	}
	b, err := json.MarshalIndent(exist, "", "  ")
	if err != nil {
		log.Println("[ERROR]", err)
		return
	}
	err = os.WriteFile(path, b, 0644)
	if err != nil {
		log.Println("[ERROR]", err)
		return
	}
}
