package model

import (
	"errors"
	"fmt"
	log2 "log"
	"os"
	"time"

	"github.com/hb1707/ant-godmin/pkg/log"
	"github.com/hb1707/ant-godmin/setting"
	"gorm.io/driver/clickhouse"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type TableBaseClickhouse struct {
	Id        string                 `json:"id" form:"id" gorm:"type:UUID;"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	DeletedAt gorm.DeletedAt         `gorm:"index" json:"-"` // 删除时间
	DB        *gorm.DB               `json:"-" form:"-" gorm:"-"`
	Req       interface{}            `json:"-" form:"-" gorm:"-"`
	Data      map[string]interface{} `json:"-" form:"-" gorm:"-"`
	Limit     int                    `json:"-" form:"-" gorm:"-"`
	Page      int                    `json:"-" form:"-" gorm:"-"`
}

// OpenClickHouse 初始化 ClickHouse 连接（按需调用，不在 init 中强制启用）
func OpenClickHouse() error {
	// 未开启或未配置则直接跳过
	chConf := setting.ClickHouse
	if !chConf.ENABLE || chConf.HOST == "" {
		// 确保在未启用或未配置时不会误用旧连接
		CHDB = nil
		return nil
	}

	var logLevel = logger.Silent
	if setting.App.RUNMODE == "dev" {
		logLevel = logger.Info
	}

	newLogger := logger.New(
		log2.New(os.Stdout, "\r\n", log2.LstdFlags),
		logger.Config{
			SlowThreshold: 2 * time.Second,
			LogLevel:      logLevel,
			Colorful:      true,
		},
	)

	protocol := "clickhouse"
	options := chConf.OPTIONS
	if options == "" {
		options = "?dial_timeout=10s&read_timeout=20s"
	}

	dsn := fmt.Sprintf("%s://%s:%s@%s:%s/%s%s",
		protocol,
		chConf.USERNAME,
		chConf.PASSWORD,
		chConf.HOST,
		chConf.PORT,
		chConf.DATABASE,
		options,
	)

	chdb, errOpen := gorm.Open(clickhouse.Open(dsn),
		&gorm.Config{
			Logger: newLogger,
			NamingStrategy: schema.NamingStrategy{
				TablePrefix:   confDB.PRE,
				SingularTable: true,
			},
		})
	if errOpen != nil {
		log.Error("ClickHouse 连接失败：", errOpen)
		CHDB = nil
		return errOpen
	}

	// 进一步校验底层连接是否可用
	sqlCh, errDB := chdb.DB()
	if errDB != nil {
		log.Error("ClickHouse 获取底层 DB 失败：", errDB)
		CHDB = nil
		return errDB
	}
	// 可选：设置连接池参数
	sqlCh.SetMaxIdleConns(5)
	sqlCh.SetMaxOpenConns(20)
	sqlCh.SetConnMaxLifetime(30 * time.Minute)

	if errPing := sqlCh.Ping(); errPing != nil {
		log.Error("ClickHouse Ping 失败：", errPing)
		_ = sqlCh.Close()
		CHDB = nil
		return errPing
	}

	CHDB = chdb
	return nil
}

// GetClickHouseDB 返回 ClickHouse 的 *gorm.DB（可能为 nil）
func GetClickHouseDB() *gorm.DB {
	return CHDB
}

func CreateTableClickHouse(dst ...interface{}) error {
	chConf := setting.ClickHouse

	// 配置未启用或未正确配置时，直接跳过
	if !chConf.ENABLE {
		return nil
	}
	if chConf.HOST == "" {
		log.Error("ClickHouse 已启用但 CH_HOST 为空，跳过 AutoMigrate")
		return nil
	}
	if !chConf.AUTOMIGRATE {
		// 未开启自动迁移，不视为错误
		return nil
	}
	if CHDB == nil {
		// 双重保险，理论上不会到这里
		log.Error("ClickHouse DB 仍为 nil，跳过 AutoMigrate")
		return nil
	}

	// 校验传入的模型，避免 nil 造成 GORM 内部 panic
	if len(dst) == 0 {
		return nil
	}
	for i, m := range dst {
		if m == nil {
			log.Error(fmt.Sprintf("ClickHouse AutoMigrate 传入第 %d 个模型为 nil，已跳过", i))
			return errors.New("nil model passed to CreateTableClickHouse")
		}
	}

	if err := CHDB.AutoMigrate(dst...); err != nil {
		log.Error("ClickHouse AutoMigrate 失败:", err)
		return err
	}
	log.Info("ClickHouse 数据库表已生成")
	return nil
}
