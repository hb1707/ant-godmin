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
	NAME           string
	RUNMODE        string
	APIURL         string
	WEBURL         string
	SHAREURL       string
	WWWURL         string
	QyWxAdminAppId string
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
var Upload struct {
	LocalPath    string
	IpfsEndpoint string
	IpfsGateway  string
}
var AliyunOSS struct {
	Endpoint        string
	AccessKeyId     string
	AccessKeySecret string
	BucketName      string
	BucketUrl       string
	BasePath        string
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
	confUpload()
	confLog()
}
func confApp() {
	app, err := Cfg.GetSection("app")
	if err != nil {
		log.Fatalf("未找到配置 'app': %v", err)
	}
	App.NAME = app.Key("APP_NAME").MustString("PDP")
	App.RUNMODE = app.Key("APP_MODE").MustString("dev")
	App.APIURL = app.Key("API_URL").MustString("")
	App.WEBURL = app.Key("WEB_URL").MustString("")
	App.WWWURL = app.Key("WWW_URL").MustString(App.WEBURL)
	App.SHAREURL = app.Key("SHARE_URL").MustString(App.WWWURL)
	App.QyWxAdminAppId = app.Key("QYWX_ADM_APPID").MustString(AdminAppid)
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
func confUpload() {
	upload, err := Cfg.GetSection("upload")
	if err == nil {
		Upload.LocalPath = "." + upload.Key("LOCAL_PATH").MustString("")
		Upload.IpfsEndpoint = upload.Key("IPFS_ENDPOINT").MustString("")
		Upload.IpfsGateway = upload.Key("IPFS_GATEWAY").MustString("")
		AliyunOSS.Endpoint = upload.Key("ALIYUN_OSS_ENDPOINT").MustString("")
		AliyunOSS.AccessKeyId = upload.Key("ALIYUN_OSS_ACCESS_KEY_ID").MustString("")
		AliyunOSS.AccessKeySecret = upload.Key("ALIYUN_OSS_ACCESS_KEY_SECRET").MustString("")
		AliyunOSS.BucketName = upload.Key("ALIYUN_OSS_BUCKET_NAME").MustString("")
		AliyunOSS.BucketUrl = upload.Key("ALIYUN_OSS_BUCKET_URL").MustString("")
		AliyunOSS.BasePath = upload.Key("ALIYUN_OSS_BASE_PATH").MustString("")
	}
}

func confLog() {
	clog, err := Cfg.GetSection("log")
	if err != nil {
		log.Fatalf("未找到配置 'log': %v", err)
	}
	Log.PATH = clog.Key("LOG_PATH").MustString("")
}
