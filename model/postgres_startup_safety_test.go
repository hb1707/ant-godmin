package model

import (
	"bytes"
	stdlog "log"
	"strings"
	"sync"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type postgresUnsafeLegacyProbe struct {
	UID  int64 `gorm:"column:uid;type:integer;not null;default:0"`
	BID  int64 `gorm:"column:bid;type:bigint;not null;default:0"`
	Flag bool  `gorm:"column:flag;type:boolean;not null;default:false"`
}

func (postgresUnsafeLegacyProbe) TableName() string { return "unsafe_legacy_probe" }

func TestPostgresStartupPreservesExistingDataAndReportsDifferences(t *testing.T) {
	db, queries := postgresMigrationTestDB(t)
	// 同时覆盖收窄、放大、跨类型和 NULL，不能只用空表证明安全。
	if err := db.Exec(`CREATE TABLE unsafe_legacy_probe(uid bigint DEFAULT 0, bid integer DEFAULT 0, flag integer DEFAULT 0);
 INSERT INTO unsafe_legacy_probe VALUES(4294967295, 1, 1), (NULL, NULL, NULL)`).Error; err != nil {
		t.Fatal(err)
	}
	var before uint32
	if err := db.Raw("SELECT relfilenode FROM pg_class WHERE oid='unsafe_legacy_probe'::regclass").Scan(&before).Error; err != nil {
		t.Fatal(err)
	}
	var reports bytes.Buffer
	oldOutput := stdlog.Writer()
	stdlog.SetOutput(&reports)
	defer stdlog.SetOutput(oldOutput)
	pool, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	for run := 0; run < 3; run++ {
		// 新建 schema 缓存，避免前次迁移修改缓存掩盖差异。共用同一专用连接保留 search_path。
		fresh, err := gorm.Open(newPostgresDialector(postgres.Config{Conn: pool}), &gorm.Config{Logger: queries})
		if err != nil {
			t.Fatal(err)
		}
		queries.ddl = nil
		if err := fresh.AutoMigrate(&postgresUnsafeLegacyProbe{}); err != nil {
			t.Fatalf("待手动迁移差异不应导致启动失败: %v", err)
		}
		if len(queries.ddl) != 0 {
			t.Fatalf("已有列不得自动改型/收紧非空: %v", queries.ddl)
		}
	}
	var after uint32
	var max string
	var nulls int64
	if err := db.Raw("SELECT relfilenode FROM pg_class WHERE oid='unsafe_legacy_probe'::regclass").Scan(&after).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Raw("SELECT max(uid)::text, count(*) FILTER (WHERE uid IS NULL) FROM unsafe_legacy_probe").Row().Scan(&max, &nulls); err != nil {
		t.Fatal(err)
	}
	if before != after || max != "4294967295" || nulls != 1 {
		t.Fatalf("已有表或数据被改写: node=%d/%d max=%s nulls=%d", before, after, max, nulls)
	}
	for _, column := range []string{"uid", "bid", "flag"} {
		if !strings.Contains(reports.String(), "column="+column+" operation=SET NOT NULL") {
			t.Fatalf("同列改型不能掩盖非空差异 %s: %s", column, reports.String())
		}
		if !strings.Contains(reports.String(), "column="+column+" operation=TYPE") {
			t.Fatalf("未报告待迁移列 %s: %s", column, reports.String())
		}
	}
}

func TestFilesOwnerIDsDeclareBigint(t *testing.T) {
	parsed, err := schema.Parse(&Files{}, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"Uid", "Bid"} {
		f := parsed.LookUpField(name)
		if string(f.DataType) != "bigint" {
			t.Fatalf("Files.%s 应与用户/分区引用的 bigint 约定一致: %s", name, f.DataType)
		}
	}
}

func TestPostgresNegativeBigintDefaultDoesNotRepeatDDL(t *testing.T) {
	db, queries := postgresMigrationTestDB(t)
	type counter struct {
		Value int64 `gorm:"type:bigint;default:-1"`
	}
	if err := db.Table("negative_counter").AutoMigrate(&counter{}); err != nil {
		t.Fatal(err)
	}
	for run := 0; run < 3; run++ {
		queries.ddl = nil
		if err := db.Table("negative_counter").AutoMigrate(&counter{}); err != nil {
			t.Fatal(err)
		}
		if len(queries.ddl) != 0 {
			t.Fatalf("catalog 的 '-1'::integer 不应反复改 bigint 默认值: %v", queries.ddl)
		}
	}
}
