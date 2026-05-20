package setting

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

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
	IsVPC     bool
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
	Region          string
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
	Region          string
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
func GetCfg() *ini.File {
	return Cfg
}
func readENV() {
	App.RUNMODE = os.Getenv("RUN_MODE")
}

func envKey(sectionName, key string) string {
	if sectionName == "" {
		return key
	}
	normalized := strings.ToUpper(sectionName)
	normalized = strings.ReplaceAll(normalized, "-", "_")
	normalized = strings.ReplaceAll(normalized, ".", "_")
	return normalized + "_" + key
}

func getString(section *ini.Section, sectionName, key, def string) string {
	if section != nil && section.HasKey(key) {
		value := strings.TrimSpace(section.Key(key).String())
		if value != "" {
			return value
		}
	}
	if value, ok := os.LookupEnv(envKey(sectionName, key)); ok && strings.TrimSpace(value) != "" {
		return value
	}
	if value, ok := os.LookupEnv(key); ok && strings.TrimSpace(value) != "" {
		return value
	}
	return def
}

func getBool(section *ini.Section, sectionName, key string, def bool) bool {
	if section != nil && section.HasKey(key) {
		raw := strings.TrimSpace(section.Key(key).String())
		if raw != "" {
			if value, err := strconv.ParseBool(raw); err == nil {
				return value
			}
		}
	}
	if raw, ok := os.LookupEnv(envKey(sectionName, key)); ok && strings.TrimSpace(raw) != "" {
		if value, err := strconv.ParseBool(raw); err == nil {
			return value
		}
	}
	if raw, ok := os.LookupEnv(key); ok && strings.TrimSpace(raw) != "" {
		if value, err := strconv.ParseBool(raw); err == nil {
			return value
		}
	}
	return def
}

func getInt(section *ini.Section, sectionName, key string, def int) int {
	if section != nil && section.HasKey(key) {
		raw := strings.TrimSpace(section.Key(key).String())
		if raw != "" {
			if value, err := strconv.Atoi(raw); err == nil {
				return value
			}
		}
	}
	if raw, ok := os.LookupEnv(envKey(sectionName, key)); ok && strings.TrimSpace(raw) != "" {
		if value, err := strconv.Atoi(strings.TrimSpace(raw)); err == nil {
			return value
		}
	}
	if raw, ok := os.LookupEnv(key); ok && strings.TrimSpace(raw) != "" {
		if value, err := strconv.Atoi(strings.TrimSpace(raw)); err == nil {
			return value
		}
	}
	return def
}

func confApp() {
	app, _ := Cfg.GetSection("app")
	App.NAME = getString(app, "app", "APP_NAME", "PDP")
	App.APIURL = getString(app, "app", "API_URL", "")
	App.WEBURL = getString(app, "app", "WEB_URL", "")
	App.WWWURL = getString(app, "app", "WWW_URL", App.WEBURL)
	App.SHAREURL = getString(app, "app", "SHARE_URL", App.WWWURL)
	App.STATICURL = getString(app, "app", "STATIC_URL", App.WEBURL)
	App.AuthKey = getString(app, "app", "AUTH_KEY", "")
	App.AesKey = getString(app, "app", "AES_KEY", "")
	if App.RUNMODE == "" {
		App.RUNMODE = getString(app, "app", "APP_MODE", "dev")
	}
	App.IsVPC = getBool(app, "app", "IS_VPC", false)
}
func confDB() {
	database, _ := GetCfg().GetSection("database")
	DB.DRIVER = getString(database, "database", "DB_DRIVER", "mysql")
	DB.HOST = getString(database, "database", "DB_HOST", "")
	DB.PORT = getString(database, "database", "DB_PORT", "")
	DB.DATABASE = getString(database, "database", "DB_DATABASE", "")
	DB.USERNAME = getString(database, "database", "DB_USERNAME", "")
	DB.PASSWORD = getString(database, "database", "DB_PASSWORD", "")
	DB.PRE = getString(database, "database", "DB_PRE", "")
	DB.AUTOMIGRATE = getBool(database, "database", "DB_AUTO_MIGRATE", false)
}
func confUpload() {
	upload, _ := Cfg.GetSection("upload")
	Upload.LocalPath = "." + getString(upload, "upload", "LOCAL_PATH", "")
	Upload.UserPath = "." + getString(upload, "upload", "USER_PATH", "")
	IPFS.IpfsEndpoint = getString(upload, "upload", "IPFS_ENDPOINT", "")
	IPFS.IpfsGateway = getString(upload, "upload", "IPFS_GATEWAY", "")
	AliyunOSS.Endpoint = getString(upload, "upload", "ALIYUN_OSS_ENDPOINT", "")
	AliyunOSS.Region = getString(upload, "upload", "ALIYUN_OSS_REGION", "cn-hangzhou")
	AliyunOSS.AccessKeyId = getString(upload, "upload", "ALIYUN_OSS_ACCESS_KEY_ID", "")
	AliyunOSS.AccessKeySecret = getString(upload, "upload", "ALIYUN_OSS_ACCESS_KEY_SECRET", "")
	AliyunOSS.BucketName = getString(upload, "upload", "ALIYUN_OSS_BUCKET_NAME", "")
	AliyunOSS.BucketNameUser = getString(upload, "upload", "ALIYUN_OSS_BUCKET_NAME_USER", "")
	AliyunOSS.BucketUrl = getString(upload, "upload", "ALIYUN_OSS_BUCKET_URL", "")
	AliyunOSS.BucketUrlUser = getString(upload, "upload", "ALIYUN_OSS_BUCKET_URL_USER", "")
	AliyunOSS.BasePath = getString(upload, "upload", "ALIYUN_OSS_BASE_PATH", "")
	AliyunOSS.MncTopic = getString(upload, "upload", "ALIYUN_MNC_TOPIC", "")

	uploadEnc, _ := Cfg.GetSection("upload_encryption")
	AliyunOSSEnc.Endpoint = getString(uploadEnc, "upload_encryption", "ALIYUN_OSS_ENDPOINT", "")
	AliyunOSSEnc.Region = getString(uploadEnc, "upload_encryption", "ALIYUN_OSS_REGION", "cn-hangzhou")
	AliyunOSSEnc.AccessKeyId = getString(uploadEnc, "upload_encryption", "ALIYUN_OSS_ACCESS_KEY_ID", "")
	AliyunOSSEnc.AccessKeySecret = getString(uploadEnc, "upload_encryption", "ALIYUN_OSS_ACCESS_KEY_SECRET", "")
	AliyunOSSEnc.BucketName = getString(uploadEnc, "upload_encryption", "ALIYUN_OSS_BUCKET_NAME", "")
	AliyunOSSEnc.BucketNameUser = getString(uploadEnc, "upload_encryption", "ALIYUN_OSS_BUCKET_NAME_USER", "")
	AliyunOSSEnc.BucketUrl = getString(uploadEnc, "upload_encryption", "ALIYUN_OSS_BUCKET_URL", "")
	AliyunOSSEnc.BucketUrlUser = getString(uploadEnc, "upload_encryption", "ALIYUN_OSS_BUCKET_URL_USER", "")
	AliyunOSSEnc.BasePath = getString(uploadEnc, "upload_encryption", "ALIYUN_OSS_BASE_PATH", "")
	AliyunOSSEnc.MncTopic = getString(uploadEnc, "upload_encryption", "ALIYUN_MNC_TOPIC", "")
}

func confTencentYun() {
	tx, _ := Cfg.GetSection("txyun")
	TencentYun.SecretId = getString(tx, "txyun", "SECRET_ID", "")
	TencentYun.SecretKey = getString(tx, "txyun", "SECRET_KEY", "")
}
func confAliYun() {
	tx, _ := Cfg.GetSection("aliyun")
	AliYun.SecretId = getString(tx, "aliyun", "SECRET_ID", "")
	AliYun.SecretKey = getString(tx, "aliyun", "SECRET_KEY", "")
	AliYun.SecretIdSMS = getString(tx, "aliyun", "SECRET_ID_S", "")
	AliYun.SecretKeySMS = getString(tx, "aliyun", "SECRET_KEY_S", "")
}

func confEmail() {
	tx, _ := Cfg.GetSection("email")
	Email.PWD = getString(tx, "email", "MAIL_SYS_PWD", "")
	Email.AdminEmail = getString(tx, "email", "MAIL_ADMIN", "")
	Email.SystemMail = getString(tx, "email", "MAIL_SYS", "")
}
func confLog() {
	clog, _ := Cfg.GetSection("log")
	Log.PATH = getString(clog, "log", "LOG_PATH", "")
	if Log.PATH == "" {
		log.Println("[WARN] 未配置 LOG_PATH，将使用默认输出")
	}
}

// ClickHouse 配置读取（可选，不存在 section 时不致命）
func confClickHouse() {
	section, _ := Cfg.GetSection("clickhouse")
	ClickHouse.ENABLE = getBool(section, "clickhouse", "CH_ENABLE", false)
	ClickHouse.HOST = getString(section, "clickhouse", "CH_HOST", "")
	ClickHouse.PORT = getString(section, "clickhouse", "CH_PORT", "9000")
	ClickHouse.DATABASE = getString(section, "clickhouse", "CH_DATABASE", "default")
	ClickHouse.USERNAME = getString(section, "clickhouse", "CH_USERNAME", "default")
	ClickHouse.PASSWORD = getString(section, "clickhouse", "CH_PASSWORD", "")
	ClickHouse.OPTIONS = getString(section, "clickhouse", "CH_OPTIONS", "")
	ClickHouse.AUTOMIGRATE = getBool(section, "clickhouse", "CH_AUTO_MIGRATE", false)
}
