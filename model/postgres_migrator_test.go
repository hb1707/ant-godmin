package model

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type PostgresMigrationProbe struct {
	ID     int64     `gorm:"primaryKey"`
	Token  string    `gorm:"type:varchar(128);not null;default:'';comment:会话令牌;uniqueIndex:probe_token"`
	Active bool      `gorm:"default:true"`
	Amount float64   `gorm:"type:numeric(10,2);default:1.0"`
	Meta   []byte    `gorm:"type:jsonb;default:'{}'"`
	SeenAt time.Time `gorm:"type:timestamptz;default:CURRENT_TIMESTAMP"`
}

func (PostgresMigrationProbe) TableName() string { return "migration_probe" }

type postgresNoDefaultProbe struct {
	PostgresMigrationProbe
	Token string `gorm:"type:varchar(128);not null;comment:会话令牌;uniqueIndex:probe_token"`
}

type postgresAddedColumnProbe struct {
	PostgresMigrationProbe
	Note string `gorm:"type:varchar(32);not null;default:'new';index:probe_note"`
}

func TestPostgresDefaultMatchesPresence(t *testing.T) {
	s, err := schema.Parse(&PostgresMigrationProbe{}, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		t.Fatal(err)
	}
	f := s.LookUpField("Token")
	for _, exists := range []bool{false, true} {
		current := postgresColumnState{defaultSQL: sql.NullString{String: "''::character varying", Valid: exists}, typeSQL: "character varying"}
		if got := postgresDefaultMatches(f, current, "''"); got != exists {
			t.Fatalf("空串与不存在必须按存在位区分: exists=%v matches=%v", exists, got)
		}
	}
	if f.DefaultValue != "" || f.DefaultValueInterface != "" || !f.HasDefaultValue {
		t.Fatal("比较不能改变模型契约")
	}
}

type migrationStatements struct {
	logger.Interface
	ddl []string
}

func (l *migrationStatements) Trace(_ context.Context, _ time.Time, fc func() (string, int64), _ error) {
	query, _ := fc()
	upper := strings.ToUpper(strings.TrimSpace(query))
	if strings.HasPrefix(upper, "ALTER ") || strings.HasPrefix(upper, "COMMENT ") || strings.HasPrefix(upper, "CREATE ") {
		l.ddl = append(l.ddl, query)
	}
}

func postgresMigrationTestDB(t *testing.T) (*gorm.DB, *migrationStatements) {
	t.Helper()
	dsn := os.Getenv("ANT_GODMIN_POSTGRES_TEST_DSN")
	if dsn == "" {
		t.Skip("设置 ANT_GODMIN_POSTGRES_TEST_DSN，使用本机专用库 center_orm_repro")
	}
	cfg, err := pgconn.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	if (cfg.Host != "127.0.0.1" && cfg.Host != "localhost" && cfg.Host != "::1") || cfg.Database != "center_orm_repro" {
		t.Fatal("DDL 测试只允许本机专用库 center_orm_repro")
	}
	queries := &migrationStatements{Interface: logger.Default}
	db, err := gorm.Open(newPostgresDialector(postgres.Config{DSN: dsn}), &gorm.Config{Logger: queries})
	if err != nil {
		t.Fatal(err)
	}
	pool, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	pool.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = pool.Close() })
	namespace := "orm_migration_" + strings.ReplaceAll(time.Now().UTC().Format("20060102150405.000000000"), ".", "_")
	quoted := `"` + namespace + `"`
	if err := db.Exec("CREATE SCHEMA " + quoted).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Exec("DROP SCHEMA " + quoted + " CASCADE").Error; err != nil {
			t.Error(err)
		}
	})
	if err := db.Exec("SET search_path TO " + quoted).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&PostgresMigrationProbe{}); err != nil {
		t.Fatal(err)
	}
	queries.ddl = nil
	return db, queries
}

func addMigrationDependency(t *testing.T, db *gorm.DB) {
	t.Helper()
	// 与故障相同的列依赖；无需引入文件清理业务和任何共享表。
	for _, query := range []string{
		`CREATE FUNCTION guard() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RETURN NEW; END $$`,
		`CREATE TRIGGER knowledge_file_reference_guard BEFORE INSERT OR UPDATE OF token ON migration_probe FOR EACH ROW EXECUTE FUNCTION guard()`,
	} {
		if err := db.Exec(query).Error; err != nil {
			t.Fatal(err)
		}
	}
}

func TestPostgresMetadataMigrationDoesNotChangeType(t *testing.T) {
	for _, trigger := range []bool{false, true} {
		for _, change := range []string{"none", "missing_default", "changed_default", "drop_default", "comment", "not_null"} {
			name := change + "/without_trigger"
			if trigger {
				name = change + "/with_trigger"
			}
			t.Run(name, func(t *testing.T) {
				db, queries := postgresMigrationTestDB(t)
				var setup string
				var target interface{} = &PostgresMigrationProbe{}
				switch change {
				case "missing_default":
					setup = "ALTER TABLE migration_probe ALTER COLUMN token DROP DEFAULT"
				case "changed_default":
					setup = "ALTER TABLE migration_probe ALTER COLUMN token SET DEFAULT 'legacy'"
				case "drop_default":
					target = &postgresNoDefaultProbe{}
				case "comment":
					setup = "COMMENT ON COLUMN migration_probe.token IS '旧注释'"
				case "not_null":
					setup = "ALTER TABLE migration_probe ALTER COLUMN token DROP NOT NULL"
				}
				if setup != "" {
					if err := db.Exec(setup).Error; err != nil {
						t.Fatal(err)
					}
				}
				if trigger {
					addMigrationDependency(t, db)
				}
				for run := 1; run <= 3; run++ {
					queries.ddl = nil
					if err := db.AutoMigrate(target); err != nil {
						t.Fatal(err)
					}
					for _, query := range queries.ddl {
						if strings.Contains(query, " TYPE ") {
							t.Fatalf("非类型差异不能触发改型: %s", query)
						}
					}
					if run > 1 || (change == "none" || change == "not_null") {
						if len(queries.ddl) != 0 {
							t.Fatalf("重复迁移没有收敛: %v", queries.ddl)
						}
					} else if len(queries.ddl) != 1 {
						t.Fatalf("应仅执行一条对应属性的 DDL: %v", queries.ddl)
					}
					t.Logf("run=%d DDL=%v", run, queries.ddl)
				}
				var actual struct {
					Default  *string
					Nullable string
				}
				if err := db.Raw(`SELECT column_default AS default, is_nullable AS nullable FROM information_schema.columns WHERE table_schema=current_schema() AND table_name='migration_probe' AND column_name='token'`).Scan(&actual).Error; err != nil {
					t.Fatal(err)
				}
				if (actual.Nullable == "YES") != (change == "not_null") || (actual.Default == nil) != (change == "drop_default") {
					t.Fatalf("真实元数据未收敛: %+v", actual)
				}
			})
		}
	}
}

func TestPostgresRealTypeDifferenceAndErrors(t *testing.T) {
	for _, trigger := range []bool{false, true} {
		name := "without_trigger"
		if trigger {
			name = "with_trigger"
		}
		t.Run(name, func(t *testing.T) {
			db, queries := postgresMigrationTestDB(t)
			if err := db.Exec("ALTER TABLE migration_probe ALTER COLUMN token TYPE varchar(64)").Error; err != nil {
				t.Fatal(err)
			}
			if trigger {
				addMigrationDependency(t, db)
			}
			queries.ddl = nil
			if err := db.AutoMigrate(&PostgresMigrationProbe{}); err != nil {
				t.Fatal(err)
			}
			if len(queries.ddl) != 0 {
				t.Fatalf("已有列改型必须延后: %v", queries.ddl)
			}
			var length int
			if err := db.Raw("SELECT character_maximum_length FROM information_schema.columns WHERE table_schema=current_schema() AND table_name='migration_probe' AND column_name='token'").Scan(&length).Error; err != nil {
				t.Fatal(err)
			}
			if length != 64 {
				t.Fatalf("启动不得改变已有列长度: %d", length)
			}
			// 延后自动迁移不是吞掉数据库错误；显式 SQL 仍必须返回真实依赖错误。
			err := db.Exec("ALTER TABLE migration_probe ALTER COLUMN token TYPE varchar(128)").Error
			if trigger {
				var pgErr *pgconn.PgError
				if !errors.As(err, &pgErr) || pgErr.Code != "0A000" {
					t.Fatalf("显式改型错误必须透传: %v", err)
				}
			} else if err != nil {
				t.Fatal(err)
			}

		})
	}
}

func TestPostgresNotNullViolationIsNotIgnored(t *testing.T) {
	db, queries := postgresMigrationTestDB(t)
	if err := db.Exec("ALTER TABLE migration_probe ALTER COLUMN token DROP NOT NULL").Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("INSERT INTO migration_probe (token) VALUES (NULL)").Error; err != nil {
		t.Fatal(err)
	}
	addMigrationDependency(t, db)
	queries.ddl = nil
	if err := db.AutoMigrate(&PostgresMigrationProbe{}); err != nil {
		t.Fatal(err)
	}
	if len(queries.ddl) != 0 {
		t.Fatalf("收紧非空只能报告，不能扫描表: %v", queries.ddl)
	}
	err := db.Exec("ALTER TABLE migration_probe ALTER COLUMN token SET NOT NULL").Error
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23502" {
		t.Fatalf("显式迁移必须暴露 NULL 数据: %v", err)
	}

}

func TestPostgresAddColumnIndexAndPrecision(t *testing.T) {
	db, queries := postgresMigrationTestDB(t)
	addMigrationDependency(t, db)
	if err := db.Exec("ALTER TABLE migration_probe ALTER COLUMN amount TYPE numeric(8,1)").Error; err != nil {
		t.Fatal(err)
	}
	queries.ddl = nil
	if err := db.AutoMigrate(&postgresAddedColumnProbe{}); err != nil {
		t.Fatal(err)
	}
	if !db.Migrator().HasColumn(&PostgresMigrationProbe{}, "note") || !db.Migrator().HasIndex(&PostgresMigrationProbe{}, "probe_note") {
		t.Fatal("正常新增字段和索引不能被迁移适配阻止")
	}
	var precision, scale int
	if err := db.Raw(`SELECT numeric_precision, numeric_scale FROM information_schema.columns WHERE table_schema=current_schema() AND table_name='migration_probe' AND column_name='amount'`).Row().Scan(&precision, &scale); err != nil {
		t.Fatal(err)
	}
	if precision != 8 || scale != 1 {
		t.Fatalf("自动迁移不应改变已有数值精度: (%d,%d)", precision, scale)
	}
	queries.ddl = nil
	if err := db.AutoMigrate(&postgresAddedColumnProbe{}); err != nil || len(queries.ddl) != 0 {
		t.Fatalf("新增字段、索引及精度迁移后未收敛: error=%v DDL=%v", err, queries.ddl)
	}
}
