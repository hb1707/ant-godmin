package model

import (
	"encoding/json"
	"gorm.io/gorm"
	"strings"
	"testing"
)

type postgresNoticeProbe struct {
	UID  int64  `gorm:"column:uid;type:bigint;not null;default:0"`
	Code string `gorm:"column:code;type:varchar(128);not null;default:''"`
}

func (postgresNoticeProbe) TableName() string { return "notice_probe" }

type postgresNoticeLogger struct {
	*migrationStatements
	batches [][]PostgresMigrationDifference
}

func (l *postgresNoticeLogger) ReportDeferredPostgresMigrations(payload []byte) {
	var differences []PostgresMigrationDifference
	if err := json.Unmarshal(payload, &differences); err != nil {
		panic(err)
	}
	l.batches = append(l.batches, differences)
}

func TestPostgresMigrationObserverBatchesRealMetadata(t *testing.T) {
	db, queries := postgresMigrationTestDB(t)
	if err := db.Exec("CREATE TABLE notice_probe(uid integer DEFAULT 0,code varchar(64) DEFAULT ''); INSERT INTO notice_probe VALUES(NULL,NULL)").Error; err != nil {
		t.Fatal(err)
	}
	observer := &postgresNoticeLogger{migrationStatements: queries}
	db = db.Session(&gorm.Session{Logger: observer})
	if err := db.Session(&gorm.Session{DryRun: true}).AutoMigrate(&postgresNoticeProbe{}); err != nil {
		t.Fatal(err)
	}
	if len(observer.batches) != 0 {
		t.Fatal("DryRun不能触发对外通知")
	}
	queries.ddl = nil
	for i := 0; i < 2; i++ {
		if err := db.AutoMigrate(&postgresNoticeProbe{}); err != nil {
			t.Fatal(err)
		}
	}
	if len(queries.ddl) != 0 || len(observer.batches) != 2 || len(observer.batches[0]) != 4 {
		t.Fatalf("应只输出逐批差异，无危险DDL: ddl=%v batches=%+v", queries.ddl, observer.batches)
	}
	seen := map[string]bool{}
	for _, d := range observer.batches[0] {
		if d.Table != "notice_probe" || d.Schema == "" {
			t.Fatalf("缺少真实结构标识: %+v", d)
		}
		seen[d.Column+"/"+d.Operation] = true
		if d.Operation == "TYPE" && d.Column == "uid" && (d.Actual != "integer" || d.Desired != "bigint") {
			t.Fatalf("UID类型应来自catalog和模型: %+v", d)
		}
		if d.Operation == "TYPE" && d.Column == "code" && !strings.Contains(d.Actual, "(64)") {
			t.Fatalf("真实长度不能丢失: %+v", d)
		}
	}
	if len(seen) != 4 {
		t.Fatalf("类型差异不能掩盖同列非空差异: %+v", seen)
	}
	var nulls int64
	if err := db.Raw("SELECT count(*) FROM notice_probe WHERE uid IS NULL AND code IS NULL").Scan(&nulls).Error; err != nil || nulls != 1 {
		t.Fatalf("数据不应被通知路径改写: %d %v", nulls, err)
	}
}
