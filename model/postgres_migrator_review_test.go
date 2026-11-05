package model

import (
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type postgresLiteralDefaultProbe struct {
	PostgresMigrationProbe
	Token string `gorm:"type:varchar(128);not null;default:can't;comment:会话令牌;uniqueIndex:probe_token"`
}

type postgresSequenceProbe struct {
	ID       int64 `gorm:"primaryKey"`
	Sequence int64 `gorm:"autoIncrement;uniqueIndex:probe_sequence"`
}

func (postgresSequenceProbe) TableName() string { return "sequence_probe" }

type postgresLargeDefaultProbe struct {
	ID      int64 `gorm:"primaryKey"`
	Counter int64 `gorm:"not null;default:9007199254740993"`
}

func (postgresLargeDefaultProbe) TableName() string { return "large_default_probe" }

type postgresAliasProbe struct {
	PostgresMigrationProbe
	Token  string  `gorm:"type:character varying(128);not null;default:'';comment:会话令牌;uniqueIndex:probe_token"`
	Amount float64 `gorm:"type:decimal(10,2);default:1.0"`
}

type postgresNullableProbe struct {
	PostgresMigrationProbe
	Token string `gorm:"type:varchar(128);default:'';comment:会话令牌;uniqueIndex:probe_token"`
}

type postgresUniqueProbe struct {
	ID   int64  `gorm:"primaryKey"`
	Code string `gorm:"type:varchar(32);unique"`
}

func (postgresUniqueProbe) TableName() string { return "unique_probe" }

type postgresFixedCharProbe struct {
	ID     int64  `gorm:"primaryKey"`
	Digest string `gorm:"type:char(64);not null;default:'';comment:'摘要'"`
}

func (postgresFixedCharProbe) TableName() string { return "fixed_char_probe" }

type postgresFunctionDefaultProbe struct {
	ID      int64 `gorm:"primaryKey"`
	Counter int64 `gorm:"default:nextval('default_counter'::regclass)"`
}

func (postgresFunctionDefaultProbe) TableName() string { return "function_default_probe" }

type postgresPreserveDefaultProbe struct {
	PostgresMigrationProbe
	Token string `gorm:"type:varchar(128);not null;default:(-);comment:会话令牌;uniqueIndex:probe_token"`
}

func TestPostgresReviewStringDefaults(t *testing.T) {
	for _, kind := range []string{"colon_literal", "quote_literal"} {
		t.Run(kind, func(t *testing.T) {
			db, queries := postgresMigrationTestDB(t)
			var target interface{} = &PostgresMigrationProbe{}
			wanted := ""
			if kind == "colon_literal" {
				if err := db.Exec("ALTER TABLE migration_probe ALTER COLUMN token SET DEFAULT '::tail'").Error; err != nil {
					t.Fatal(err)
				}
			} else {
				target = &postgresLiteralDefaultProbe{}
				wanted = "can't"
			}
			addMigrationDependency(t, db)
			for run := 1; run <= 2; run++ {
				queries.ddl = nil
				if err := db.AutoMigrate(target); err != nil {
					t.Fatal(err)
				}
				if run == 2 && len(queries.ddl) != 0 {
					t.Fatalf("字面量默认值必须收敛: %v", queries.ddl)
				}
			}
			var actual string
			if err := db.Raw("INSERT INTO migration_probe DEFAULT VALUES RETURNING token").Scan(&actual).Error; err != nil {
				t.Fatal(err)
			}
			if actual != wanted {
				t.Fatalf("真实字符串默认值差异被漏掉: wanted=%q actual=%q", wanted, actual)
			}
		})
	}
}

func TestPostgresReviewNonPrimarySequence(t *testing.T) {
	db, queries := postgresMigrationTestDB(t)
	for run := 1; run <= 2; run++ {
		queries.ddl = nil
		if err := db.AutoMigrate(&postgresSequenceProbe{}); err != nil {
			t.Fatal(err)
		}
		if run == 2 && len(queries.ddl) != 0 {
			t.Fatalf("序列字段不能重复 DDL: %v", queries.ddl)
		}
	}
	var actual int64
	if err := db.Raw("INSERT INTO sequence_probe DEFAULT VALUES RETURNING sequence").Scan(&actual).Error; err != nil {
		t.Fatal(err)
	}
	if actual != 1 {
		t.Fatalf("非主键序列生成异常: %d", actual)
	}
}

func TestPostgresReviewLargeNumericDefault(t *testing.T) {
	db, queries := postgresMigrationTestDB(t)
	if err := db.AutoMigrate(&postgresLargeDefaultProbe{}); err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("ALTER TABLE large_default_probe ALTER COLUMN counter SET DEFAULT 9007199254740992").Error; err != nil {
		t.Fatal(err)
	}
	previous := DB
	DB = db
	t.Cleanup(func() { DB = previous })
	// 完整启动路径会先调用此对齐；不能在浮点舍入后把两个不同整数视为相同。
	if err := alignPostgresExistingColumns(&postgresLargeDefaultProbe{}); err != nil {
		t.Fatal(err)
	}
	queries.ddl = nil
	if err := db.AutoMigrate(&postgresLargeDefaultProbe{}); err != nil {
		t.Fatal(err)
	}
	var actual string
	if err := db.Raw("SELECT column_default FROM information_schema.columns WHERE table_schema=current_schema() AND table_name='large_default_probe' AND column_name='counter'").Scan(&actual).Error; err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(actual, "9007199254740993") {
		t.Fatalf("不同大整数默认值被判为等价: %s DDL=%v", actual, queries.ddl)
	}
}

func TestPostgresReviewAliasesAndMixedChanges(t *testing.T) {
	db, queries := postgresMigrationTestDB(t)
	for _, query := range []string{
		"ALTER TABLE migration_probe ALTER COLUMN token TYPE varchar(64)",
		"ALTER TABLE migration_probe ALTER COLUMN token DROP DEFAULT",
		"ALTER TABLE migration_probe ALTER COLUMN token DROP NOT NULL",
		"COMMENT ON COLUMN migration_probe.token IS '旧注释'",
		"ALTER TABLE migration_probe ALTER COLUMN amount TYPE numeric(8,1)",
	} {
		if err := db.Exec(query).Error; err != nil {
			t.Fatal(err)
		}
	}
	queries.ddl = nil
	if err := db.AutoMigrate(&postgresAliasProbe{}); err != nil {
		t.Fatal(err)
	}
	var length, precision, scale int
	if err := db.Raw("SELECT character_maximum_length FROM information_schema.columns WHERE table_schema=current_schema() AND table_name='migration_probe' AND column_name='token'").Scan(&length).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Raw("SELECT numeric_precision,numeric_scale FROM information_schema.columns WHERE table_schema=current_schema() AND table_name='migration_probe' AND column_name='amount'").Row().Scan(&precision, &scale); err != nil {
		t.Fatal(err)
	}
	if length != 64 || precision != 8 || scale != 1 {
		t.Fatalf("已有长度和精度应等待显式迁移: %d (%d,%d) DDL=%v", length, precision, scale, queries.ddl)
	}
	defaultChanges, nullChanges := 0, 0
	for _, query := range queries.ddl {
		if !strings.Contains(query, `ALTER COLUMN "token"`) {
			continue
		}
		if strings.Contains(query, "SET DEFAULT") {
			defaultChanges++
		}
		if strings.Contains(query, "SET NOT NULL") {
			nullChanges++
		}
	}
	if defaultChanges != 0 || nullChanges != 0 {
		t.Fatalf("混合变更不应因旧元数据重复 DDL: %v", queries.ddl)
	}
	queries.ddl = nil
	if err := db.AutoMigrate(&postgresAliasProbe{}); err != nil || len(queries.ddl) != 0 {
		t.Fatalf("别名类型迁移不收敛: %v %v", err, queries.ddl)
	}
}

func TestPostgresReviewNullableAndUnique(t *testing.T) {
	db, queries := postgresMigrationTestDB(t)
	addMigrationDependency(t, db)
	queries.ddl = nil
	if err := db.AutoMigrate(&postgresNullableProbe{}); err != nil {
		t.Fatal(err)
	}
	if len(queries.ddl) != 1 || !strings.Contains(queries.ddl[0], "DROP NOT NULL") {
		t.Fatalf("放宽非空应独立处理: %v", queries.ddl)
	}
	if err := db.Exec("CREATE TABLE unique_probe (id bigserial PRIMARY KEY, code varchar(32))").Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("INSERT INTO unique_probe (code) VALUES ('duplicate'),('duplicate')").Error; err != nil {
		t.Fatal(err)
	}
	err := db.AutoMigrate(&postgresUniqueProbe{})
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
		t.Fatalf("添加唯一约束的真实冲突不能被吞掉: %v", err)
	}
	if err := db.Exec("DELETE FROM unique_probe WHERE id=2").Error; err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&postgresUniqueProbe{}); err != nil {
		t.Fatal(err)
	}
	if !db.Migrator().HasConstraint(&postgresUniqueProbe{}, "uni_unique_probe_code") {
		t.Fatal("唯一约束未创建")
	}
	queries.ddl = nil
	if err := db.AutoMigrate(&postgresUniqueProbe{}); err != nil || len(queries.ddl) != 0 {
		t.Fatalf("唯一约束重复迁移异常: %v %v", err, queries.ddl)
	}
}

func TestPostgresReviewFixedCharAndFunctionDefault(t *testing.T) {
	db, queries := postgresMigrationTestDB(t)
	if err := db.Exec("CREATE SEQUENCE default_counter").Error; err != nil {
		t.Fatal(err)
	}
	for run := 1; run <= 3; run++ {
		queries.ddl = nil
		if err := db.AutoMigrate(&postgresFixedCharProbe{}, &postgresFunctionDefaultProbe{}); err != nil {
			t.Fatal(err)
		}
		if run > 1 && len(queries.ddl) != 0 {
			t.Fatalf("char 默认值、带引号注释或函数默认值不收敛: %v", queries.ddl)
		}
	}
	var called bool
	if err := db.Raw("SELECT is_called FROM default_counter").Scan(&called).Error; err != nil {
		t.Fatal(err)
	}
	if called {
		t.Fatal("比较默认值不能执行 nextval 产生副作用")
	}
	if err := db.Exec("ALTER TABLE migration_probe ALTER COLUMN token SET DEFAULT 'database_owned'").Error; err != nil {
		t.Fatal(err)
	}
	queries.ddl = nil
	if err := db.AutoMigrate(&postgresPreserveDefaultProbe{}); err != nil {
		t.Fatal(err)
	}
	if len(queries.ddl) != 0 {
		t.Fatalf("default:(-) 必须保持数据库原有默认值: %v", queries.ddl)
	}
}

func TestPostgresReviewDryRunAndPreparedStatements(t *testing.T) {
	db, queries := postgresMigrationTestDB(t)
	if err := db.Exec("ALTER TABLE migration_probe ALTER COLUMN token DROP DEFAULT").Error; err != nil {
		t.Fatal(err)
	}
	queries.ddl = nil
	if err := db.Session(&gorm.Session{DryRun: true}).AutoMigrate(&PostgresMigrationProbe{}); err != nil {
		t.Fatal(err)
	}
	var hasDefault bool
	if err := db.Raw("SELECT column_default IS NOT NULL FROM information_schema.columns WHERE table_schema=current_schema() AND table_name='migration_probe' AND column_name='token'").Scan(&hasDefault).Error; err != nil {
		t.Fatal(err)
	}
	if hasDefault {
		t.Fatal("DryRun 不能实际补写默认值")
	}
	if err := db.Exec("ALTER TABLE migration_probe ALTER COLUMN token TYPE varchar(64)").Error; err != nil {
		t.Fatal(err)
	}
	prepared := db.Session(&gorm.Session{PrepareStmt: true})
	if err := prepared.AutoMigrate(&PostgresMigrationProbe{}); err != nil {
		t.Fatal(err)
	}
	queries.ddl = nil
	if err := prepared.AutoMigrate(&PostgresMigrationProbe{}); err != nil || len(queries.ddl) != 0 {
		t.Fatalf("prepared statements 下改型后重复迁移异常: %v %v", err, queries.ddl)
	}
}

func TestPostgresReviewDefaultComparisonIsExact(t *testing.T) {
	field := &schema.Field{GORMDataType: schema.Time, HasDefaultValue: true, DefaultValue: "custom_time('Admin')"}
	current := postgresColumnState{defaultSQL: sql.NullString{String: "custom_time('admin')", Valid: true}}
	if postgresDefaultMatches(field, current, field.DefaultValue) {
		t.Fatal("函数参数不能因为大小写折叠被视为等价")
	}
	for _, sample := range []struct {
		a, b  string
		equal bool
	}{
		{"1", "1.0", true}, {"1e3", "1000", true}, {"9007199254740992", "9007199254740993", false}, {"1/2", "0.5", false},
	} {
		if got := equalPostgresNumericDefaults(sample.a, sample.b); got != sample.equal {
			t.Fatalf("数值字面量比较错误: %+v result=%v", sample, got)
		}
	}
}
