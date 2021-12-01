package orm

import (
    "antGodmin/pkg/log"
    "antGodmin/setting"
    "database/sql"
    "errors"
    "fmt"
    "github.com/hb1707/exfun/fun"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
    "gorm.io/gorm/schema"
    log2 "log"
    "os"
    "time"
)

var (
    sqlDB  *sql.DB
    db     *gorm.DB
    err    error
    confDB = setting.DB
)

type DatelineMap struct {
    Str string `json:"str"`
}

type TableBase struct {
    Id        uint                   `json:"id" gorm:"type:int(10) UNSIGNED not null AUTO_INCREMENT;primaryKey;"`
    CreatedAt time.Time              `json:"created_at"`
    UpdatedAt time.Time              `json:"updated_at"`
    DeletedAt gorm.DeletedAt         `gorm:"index" json:"-"` // 删除时间
    DB        *gorm.DB               `json:"-" gorm:"-"`
    Req       interface{}            `json:"-" gorm:"-"`
    Data      map[string]interface{} `json:"-" gorm:"-"`
    Limit     int                    `json:"-" gorm:"-"`
    Page      int                    `json:"-" gorm:"-"`
}

func OpenDB() {
    var logLevel = logger.Silent
    if setting.App.RUNMODE == "dev" {
        logLevel = logger.Info
    }

    newLogger := logger.New(
        log2.New(os.Stdout, "\r\n", log2.LstdFlags), // io writer
        logger.Config{
            SlowThreshold: time.Second, // Slow SQL threshold
            LogLevel:      logLevel,    // Log level
            Colorful:      true,        // Disable color
        },
    )

    dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
        confDB.USERNAME,
        confDB.PASSWORD,
        confDB.HOST,
        confDB.PORT,
        confDB.DATABASE)
    db, err = gorm.Open(mysql.New(mysql.Config{
        DSN:                       dsn,   // DSN data source name
        DefaultStringSize:         256,   // string 类型字段的默认长度
        DisableDatetimePrecision:  true,  // 禁用 datetime 精度，MySQL 5.6 之前的数据库不支持
        DontSupportRenameIndex:    true,  // 重命名索引时采用删除并新建的方式，MySQL 5.7 之前的数据库和 MariaDB 不支持重命名索引
        DontSupportRenameColumn:   true,  // 用 `change` 重命名列，MySQL 8 之前的数据库和 MariaDB 不支持重命名列
        SkipInitializeWithVersion: false, // 根据当前 MySQL 版本自动配置
    }), &gorm.Config{Logger: newLogger, NamingStrategy: schema.NamingStrategy{
        TablePrefix:   confDB.PRE,
        SingularTable: true,
    }})

    if err != nil {
        log.Fatal(err, 3)
    }
    sqlDB, err = db.DB()
    if err != nil {
        log.Fatal(err, 3)
    }
    sqlDB.SetMaxIdleConns(10)
    sqlDB.SetMaxOpenConns(100)
    CreatTable()
}
func CreatTable() {
    var (
        question       = PdpQuestion{}
        evaluation     = PdpEvaluation{}
        questionOption = PdpQuestionOption{}
        user           = User{}
        userScore      = UserScore{}
        userAnswer     = PdpUserAnswer{}
        userComments   = UserComments{}
        course         = Course{}
        courseHour     = CourseHours{}
        sms            = Sms{}
        autoTask       = AutoTask{}
    )
    if confDB.PRE != "" && confDB.AUTOMIGRATE {
        err := db.AutoMigrate(
            &question,
            &evaluation,
            &questionOption,
            &user,
            &userScore,
            &userAnswer,
            &userComments,
            &course,
            &courseHour,
            &sms,
            &autoTask,
        )
        if err != nil {
            log.Fatal(err)
        }
    }
}
func CloseDB() {
    err = sqlDB.Close()
    if err != nil {
        log.Error("数据库连接关闭出错了！")
    }
}

type SqlErrType int

const (
    ErrNil            SqlErrType = 0
    ErrSql            SqlErrType = 1
    ErrRecordNotFound SqlErrType = 2
    ErrExist          SqlErrType = 3
)

func failed(err error) bool {
    return failedType(err) > ErrNil
}

func failedType(err error) SqlErrType {
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return ErrRecordNotFound
        } else if fun.Stripos(err.Error(), "UNIQUE constraint failed") > -1 || fun.Stripos(err.Error(), "Error 1062: Duplicate entry") > -1 {
            return ErrExist
        }
        log.Error(err)
        return ErrSql
    } else {
        return ErrNil
    }
}
