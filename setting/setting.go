package setting

import (
	"fmt"
	"log"
	"os"

	"github.com/go-ini/ini"
)

var (
	Cfg *ini.File
)

var IsTest = false
var IsCMS = false
var IsReg = false

var App struct {
	NAME      string
	RUNMODE   string
	APIURL    string
	WEBURL    string
	SHAREURL  string
	WWWURL    string
	STATICURL string
	AuthKey   string
	AesKey    string
}

var DB struct {
	DRIVER      string
	HOST        string
	PORT        string
	DATABASE    string
	USERNAME    string
	PASSWORD    string
	PRE         string
	AUTOMIGRATE bool
}
var Upload struct {
	LocalPath string
	UserPath  string
}

var IPFS struct {
	IpfsEndpoint string
	IpfsGateway  string
}

var AliyunOSS struct {
	Endpoint        string
	AccessKeyId     string
	AccessKeySecret string
	BucketName      string
	BucketNameUser  string
	BucketUrl       string
	BucketUrlUser   string
	BasePath        string
	MncTopic        string
}
var AliyunOSSEnc struct {
	Endpoint        string
	AccessKeyId     string
	AccessKeySecret string
	BucketName      string
	BucketNameUser  string
	BucketUrl       string
	BucketUrlUser   string
	BasePath        string
	MncTopic        string
}

var Email struct {
	PWD        string
	SystemMail string
	AdminEmail string
}

var TencentYun struct {
	SecretId  string
	SecretKey string
}
var AliYun struct {
	SecretId     string
	SecretKey    string
	SecretIdSMS  string
	SecretKeySMS string
}
var Log struct {
	PATH string
}

// ClickHouse 数据库配置
var ClickHouse struct {
	ENABLE      bool
	HOST        string
	PORT        string
	DATABASE    string
	USERNAME    string
	PASSWORD    string
	OPTIONS     string
	AUTOMIGRATE bool
}

func InitConf(path string) {
	var err error
	var envPath = path + ".env"
	if os.Getenv("APP_ENV") == "dev" {
		fmt.Println("DEV模式开启")
		envPath = path + ".env.dev"
		IsTest = true
	}
	Cfg, err = ini.Load(envPath)
	if err != nil {
		fmt.Printf("找不到配置文件: %v", err)
		os.Exit(1)
	}
	readENV()
	confApp()
	confDB()
	confRedis()
	confUpload()
	confLog()
	confTencentYun()
	confAliYun()
	confEmail()
	confQyWxAdmin()
	confCoze()
	confVolc()
	confDify()
	confClickHouse() // 新增：加载 ClickHouse 配置

}

func readENV() {
	App.RUNMODE = os.Getenv("RUN_MODE")
}

func confApp() {
	app, err := Cfg.GetSection("app")
	if err != nil {
		log.Fatalf("未找到配置 'app': %v", err)
	}
	App.NAME = app.Key("APP_NAME").MustString("PDP")
	App.APIURL = app.Key("API_URL").MustString("")
	App.WEBURL = app.Key("WEB_URL").MustString("")
	App.WWWURL = app.Key("WWW_URL").MustString(App.WEBURL)
	App.SHAREURL = app.Key("SHARE_URL").MustString(App.WWWURL)
	App.STATICURL = app.Key("STATIC_URL").MustString(App.WEBURL)
	App.AuthKey = app.Key("AUTH_KEY").MustString("")
	App.AesKey = app.Key("AES_KEY").MustString("")
	if App.RUNMODE == "" {
		App.RUNMODE = app.Key("APP_MODE").MustString("dev")
	}
}
func confDB() {
	database, err := Cfg.GetSection("database")
	if err != nil {
		log.Fatalf("未找到配置 'database': %v", err)
	}
	DB.DRIVER = database.Key("DB_DRIVER").MustString("mysql")
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
		Upload.UserPath = "." + upload.Key("USER_PATH").MustString("")
		IPFS.IpfsEndpoint = upload.Key("IPFS_ENDPOINT").MustString("")
		IPFS.IpfsGateway = upload.Key("IPFS_GATEWAY").MustString("")
		AliyunOSS.Endpoint = upload.Key("ALIYUN_OSS_ENDPOINT").MustString("")
		AliyunOSS.AccessKeyId = upload.Key("ALIYUN_OSS_ACCESS_KEY_ID").MustString("")
		AliyunOSS.AccessKeySecret = upload.Key("ALIYUN_OSS_ACCESS_KEY_SECRET").MustString("")
		AliyunOSS.BucketName = upload.Key("ALIYUN_OSS_BUCKET_NAME").MustString("")
		AliyunOSS.BucketNameUser = upload.Key("ALIYUN_OSS_BUCKET_NAME_USER").MustString("")
		AliyunOSS.BucketUrl = upload.Key("ALIYUN_OSS_BUCKET_URL").MustString("")
		AliyunOSS.BucketUrlUser = upload.Key("ALIYUN_OSS_BUCKET_URL_USER").MustString("")
		AliyunOSS.BasePath = upload.Key("ALIYUN_OSS_BASE_PATH").MustString("")
		AliyunOSS.MncTopic = upload.Key("ALIYUN_MNC_TOPIC").MustString("")
	}
	uploadEnc, err := Cfg.GetSection("upload_encryption")
	if err == nil {
		AliyunOSSEnc.Endpoint = uploadEnc.Key("ALIYUN_OSS_ENDPOINT").MustString("")
		AliyunOSSEnc.AccessKeyId = uploadEnc.Key("ALIYUN_OSS_ACCESS_KEY_ID").MustString("")
		AliyunOSSEnc.AccessKeySecret = uploadEnc.Key("ALIYUN_OSS_ACCESS_KEY_SECRET").MustString("")
		AliyunOSSEnc.BucketName = uploadEnc.Key("ALIYUN_OSS_BUCKET_NAME").MustString("")
		AliyunOSSEnc.BucketNameUser = uploadEnc.Key("ALIYUN_OSS_BUCKET_NAME_USER").MustString("")
		AliyunOSSEnc.BucketUrl = uploadEnc.Key("ALIYUN_OSS_BUCKET_URL").MustString("")
		AliyunOSSEnc.BucketUrlUser = uploadEnc.Key("ALIYUN_OSS_BUCKET_URL_USER").MustString("")
		AliyunOSSEnc.BasePath = uploadEnc.Key("ALIYUN_OSS_BASE_PATH").MustString("")
		AliyunOSSEnc.MncTopic = uploadEnc.Key("ALIYUN_MNC_TOPIC").MustString("")
	}
}

func confTencentYun() {
	tx, err := Cfg.GetSection("txyun")
	if err == nil {
		TencentYun.SecretId = tx.Key("SECRET_ID").MustString("")
		TencentYun.SecretKey = tx.Key("SECRET_KEY").MustString("")
	}
}
func confAliYun() {
	tx, err := Cfg.GetSection("aliyun")
	if err == nil {
		AliYun.SecretId = tx.Key("SECRET_ID").MustString("")
		AliYun.SecretKey = tx.Key("SECRET_KEY").MustString("")
		AliYun.SecretIdSMS = tx.Key("SECRET_ID_S").MustString("")
		AliYun.SecretKeySMS = tx.Key("SECRET_KEY_S").MustString("")
	}
}

func confEmail() {
	tx, err := Cfg.GetSection("email")
	if err == nil {
		Email.PWD = tx.Key("MAIL_SYS_PWD").MustString("")
		Email.AdminEmail = tx.Key("MAIL_ADMIN").MustString("")
		Email.SystemMail = tx.Key("MAIL_SYS").MustString("")
	}
}
func confLog() {
	clog, err := Cfg.GetSection("log")
	if err != nil {
		log.Fatalf("未找到配置 'log': %v", err)
	}
	Log.PATH = clog.Key("LOG_PATH").MustString("")
}

// ClickHouse 配置读取（可选，不存在 section 时不致命）
func confClickHouse() {
	section, err := Cfg.GetSection("clickhouse")
	if err != nil {
		// 未配置 ClickHouse，直接返回
		return
	}
	ClickHouse.ENABLE = section.Key("CH_ENABLE").MustBool(false)
	ClickHouse.HOST = section.Key("CH_HOST").MustString("")
	ClickHouse.PORT = section.Key("CH_PORT").MustString("9000")
	ClickHouse.DATABASE = section.Key("CH_DATABASE").MustString("default")
	ClickHouse.USERNAME = section.Key("CH_USERNAME").MustString("default")
	ClickHouse.PASSWORD = section.Key("CH_PASSWORD").MustString("")
	ClickHouse.OPTIONS = section.Key("CH_OPTIONS").MustString("")
	ClickHouse.AUTOMIGRATE = section.Key("CH_AUTO_MIGRATE").MustBool(false)
}
