package setting

import (
	"fmt"
	"github.com/go-ini/ini"
	"log"
	"os"
)

var (
	Cfg *ini.File
)

var App struct {
	NAME    string
	RUNMODE string
	URL     string
}

var DB struct {
	HOST        string
	PORT        string
	DATABASE    string
	USERNAME    string
	PASSWORD    string
	PRE         string
	AUTOMIGRATE bool
}

var Log struct {
	PATH string
}

func init() {
	var err error
	Cfg, err = ini.Load("./config/.env")
	if err != nil {
		fmt.Printf("找不到配置文件: %v", err)
		os.Exit(1)
	}
	confApp()
	confDB()
	confLog()
}
func confApp() {
	app, err := Cfg.GetSection("app")
	if err != nil {
		log.Fatalf("未找到配置 'app': %v", err)
	}
	App.NAME = app.Key("APP_NAME").MustString("PDP")
	App.RUNMODE = app.Key("APP_MODE").MustString("dev")
	App.URL = app.Key("APP_URL").MustString("")
}
func confDB() {
	database, err := Cfg.GetSection("database")
	if err != nil {
		log.Fatalf("未找到配置 'database': %v", err)
	}
	DB.HOST = database.Key("DB_HOST").MustString("")
	DB.PORT = database.Key("DB_PORT").MustString("")
	DB.DATABASE = database.Key("DB_DATABASE").MustString("")
	DB.USERNAME = database.Key("DB_USERNAME").MustString("")
	DB.PASSWORD = database.Key("DB_PASSWORD").MustString("")
	DB.PRE = database.Key("DB_PRE").MustString("")
	DB.AUTOMIGRATE = database.Key("DB_AUTO_MIGRATE").MustBool(false)
}

func confLog() {
	clog, err := Cfg.GetSection("log")
	if err != nil {
		log.Fatalf("未找到配置 'log': %v", err)
	}
	Log.PATH = clog.Key("LOG_PATH").MustString("")
}
