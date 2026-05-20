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

func GetString(section *ini.Section, sectionName, key, def string) string {
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

func GetBool(section *ini.Section, sectionName, key string, def bool) bool {
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

func GetInt(section *ini.Section, sectionName, key string, def int) int {
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

func GetFloat(section *ini.Section, sectionName, key string, def float64) float64 {
	if section != nil && section.HasKey(key) {
		raw := strings.TrimSpace(section.Key(key).String())
		if raw != "" {
			if value, err := strconv.ParseFloat(raw, 64); err == nil {
				return value
			}
		}
	}
	if raw, ok := os.LookupEnv(envKey(sectionName, key)); ok && strings.TrimSpace(raw) != "" {
		if value, err := strconv.ParseFloat(strings.TrimSpace(raw), 64); err == nil {
			return value
		}
	}
	if raw, ok := os.LookupEnv(key); ok && strings.TrimSpace(raw) != "" {
		if value, err := strconv.ParseFloat(strings.TrimSpace(raw), 64); err == nil {
			return value
		}
	}
	return def
}

func confApp() {
	app, _ := Cfg.GetSection("app")
	App.NAME = GetString(app, "app", "APP_NAME", "PDP")
	App.APIURL = GetString(app, "app", "API_URL", "")
	App.WEBURL = GetString(app, "app", "WEB_URL", "")
	App.WWWURL = GetString(app, "app", "WWW_URL", App.WEBURL)
	App.SHAREURL = GetString(app, "app", "SHARE_URL", App.WWWURL)
	App.STATICURL = GetString(app, "app", "STATIC_URL", App.WEBURL)
	App.AuthKey = GetString(app, "app", "AUTH_KEY", "")
	App.AesKey = GetString(app, "app", "AES_KEY", "")
	if App.RUNMODE == "" {
		App.RUNMODE = GetString(app, "app", "APP_MODE", "dev")
	}
	App.IsVPC = GetBool(app, "app", "IS_VPC", false)
}
func confDB() {
	database, _ := GetCfg().GetSection("database")
	DB.DRIVER = GetString(database, "database", "DB_DRIVER", "mysql")
	DB.HOST = GetString(database, "database", "DB_HOST", "")
	DB.PORT = GetString(database, "database", "DB_PORT", "")
	DB.DATABASE = GetString(database, "database", "DB_DATABASE", "")
	DB.USERNAME = GetString(database, "database", "DB_USERNAME", "")
	DB.PASSWORD = GetString(database, "database", "DB_PASSWORD", "")
	DB.PRE = GetString(database, "database", "DB_PRE", "")
	DB.AUTOMIGRATE = GetBool(database, "database", "DB_AUTO_MIGRATE", false)
}
func confUpload() {
	upload, _ := Cfg.GetSection("upload")
	Upload.LocalPath = "." + GetString(upload, "upload", "LOCAL_PATH", "")
	Upload.UserPath = "." + GetString(upload, "upload", "USER_PATH", "")
	IPFS.IpfsEndpoint = GetString(upload, "upload", "IPFS_ENDPOINT", "")
	IPFS.IpfsGateway = GetString(upload, "upload", "IPFS_GATEWAY", "")
	AliyunOSS.Endpoint = GetString(upload, "upload", "ALIYUN_OSS_ENDPOINT", "")
	AliyunOSS.Region = GetString(upload, "upload", "ALIYUN_OSS_REGION", "cn-hangzhou")
	AliyunOSS.AccessKeyId = GetString(upload, "upload", "ALIYUN_OSS_ACCESS_KEY_ID", "")
	AliyunOSS.AccessKeySecret = GetString(upload, "upload", "ALIYUN_OSS_ACCESS_KEY_SECRET", "")
	AliyunOSS.BucketName = GetString(upload, "upload", "ALIYUN_OSS_BUCKET_NAME", "")
	AliyunOSS.BucketNameUser = GetString(upload, "upload", "ALIYUN_OSS_BUCKET_NAME_USER", "")
	AliyunOSS.BucketUrl = GetString(upload, "upload", "ALIYUN_OSS_BUCKET_URL", "")
	AliyunOSS.BucketUrlUser = GetString(upload, "upload", "ALIYUN_OSS_BUCKET_URL_USER", "")
	AliyunOSS.BasePath = GetString(upload, "upload", "ALIYUN_OSS_BASE_PATH", "")
	AliyunOSS.MncTopic = GetString(upload, "upload", "ALIYUN_MNC_TOPIC", "")

	uploadEnc, _ := Cfg.GetSection("upload_encryption")
	AliyunOSSEnc.Endpoint = GetString(uploadEnc, "upload_encryption", "ALIYUN_OSS_ENDPOINT", "")
	AliyunOSSEnc.Region = GetString(uploadEnc, "upload_encryption", "ALIYUN_OSS_REGION", "cn-hangzhou")
	AliyunOSSEnc.AccessKeyId = GetString(uploadEnc, "upload_encryption", "ALIYUN_OSS_ACCESS_KEY_ID", "")
	AliyunOSSEnc.AccessKeySecret = GetString(uploadEnc, "upload_encryption", "ALIYUN_OSS_ACCESS_KEY_SECRET", "")
	AliyunOSSEnc.BucketName = GetString(uploadEnc, "upload_encryption", "ALIYUN_OSS_BUCKET_NAME", "")
	AliyunOSSEnc.BucketNameUser = GetString(uploadEnc, "upload_encryption", "ALIYUN_OSS_BUCKET_NAME_USER", "")
	AliyunOSSEnc.BucketUrl = GetString(uploadEnc, "upload_encryption", "ALIYUN_OSS_BUCKET_URL", "")
	AliyunOSSEnc.BucketUrlUser = GetString(uploadEnc, "upload_encryption", "ALIYUN_OSS_BUCKET_URL_USER", "")
	AliyunOSSEnc.BasePath = GetString(uploadEnc, "upload_encryption", "ALIYUN_OSS_BASE_PATH", "")
	AliyunOSSEnc.MncTopic = GetString(uploadEnc, "upload_encryption", "ALIYUN_MNC_TOPIC", "")
}

func confTencentYun() {
	tx, _ := Cfg.GetSection("txyun")
	TencentYun.SecretId = GetString(tx, "txyun", "SECRET_ID", "")
	TencentYun.SecretKey = GetString(tx, "txyun", "SECRET_KEY", "")
}
func confAliYun() {
	tx, _ := Cfg.GetSection("aliyun")
	AliYun.SecretId = GetString(tx, "aliyun", "SECRET_ID", "")
	AliYun.SecretKey = GetString(tx, "aliyun", "SECRET_KEY", "")
	AliYun.SecretIdSMS = GetString(tx, "aliyun", "SECRET_ID_S", "")
	AliYun.SecretKeySMS = GetString(tx, "aliyun", "SECRET_KEY_S", "")
}

func confEmail() {
	tx, _ := Cfg.GetSection("email")
	Email.PWD = GetString(tx, "email", "MAIL_SYS_PWD", "")
	Email.AdminEmail = GetString(tx, "email", "MAIL_ADMIN", "")
	Email.SystemMail = GetString(tx, "email", "MAIL_SYS", "")
}
func confLog() {
	clog, _ := Cfg.GetSection("log")
	Log.PATH = GetString(clog, "log", "LOG_PATH", "")
	if Log.PATH == "" {
		log.Println("[WARN] 未配置 LOG_PATH，将使用默认输出")
	}
}

// ClickHouse 配置读取（可选，不存在 section 时不致命）
func confClickHouse() {
	section, _ := Cfg.GetSection("clickhouse")
	ClickHouse.ENABLE = GetBool(section, "clickhouse", "CH_ENABLE", false)
	ClickHouse.HOST = GetString(section, "clickhouse", "CH_HOST", "")
	ClickHouse.PORT = GetString(section, "clickhouse", "CH_PORT", "9000")
	ClickHouse.DATABASE = GetString(section, "clickhouse", "CH_DATABASE", "default")
	ClickHouse.USERNAME = GetString(section, "clickhouse", "CH_USERNAME", "default")
	ClickHouse.PASSWORD = GetString(section, "clickhouse", "CH_PASSWORD", "")
	ClickHouse.OPTIONS = GetString(section, "clickhouse", "CH_OPTIONS", "")
	ClickHouse.AUTOMIGRATE = GetBool(section, "clickhouse", "CH_AUTO_MIGRATE", false)
}
