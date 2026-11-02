package model

import (
	"fmt"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/clickhouse"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func clickHouseMigrationTestDB(t *testing.T) (*gorm.DB, *migrationStatements) {
	t.Helper()
	dsn := os.Getenv("ANT_GODMIN_CLICKHOUSE_TEST_DSN")
	if dsn == "" {
		t.Skip("设置 ANT_GODMIN_CLICKHOUSE_TEST_DSN，使用本机专用库 center_orm_repro")
	}
	u, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	if (u.Hostname() != "127.0.0.1" && u.Hostname() != "localhost" && u.Hostname() != "::1") || u.Path != "/center_orm_repro" {
		t.Fatal("DDL 测试只允许本机专用库 center_orm_repro")
	}
	queries := &migrationStatements{Interface: logger.Default}
	db, err := gorm.Open(clickHouseDialector{clickhouse.Open(dsn).(*clickhouse.Dialector)}, &gorm.Config{Logger: queries})
	if err != nil {
		t.Fatal(err)
	}
	p, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = p.Close() })
	return db, queries
}
func chExec(t *testing.T, db *gorm.DB, q string) {
	t.Helper()
	if err := db.Exec(q).Error; err != nil {
		t.Fatal(err)
	}
}
func chTable(t *testing.T, db *gorm.DB) string {
	t.Helper()
	name := fmt.Sprintf("ch_contract_%d", time.Now().UnixNano())
	t.Cleanup(func() {
		if err := db.Exec("DROP TABLE IF EXISTS `" + name + "`").Error; err != nil {
			t.Error(err)
		}
	})
	return name
}
func chConverges(t *testing.T, db *gorm.DB, q *migrationStatements, name string, value interface{}) {
	t.Helper()
	// 每次重新打开连接及 schema 缓存，不能依赖旧迁移在内存中留下的状态。
	for i := 0; i < 3; i++ {
		fresh, queries := clickHouseMigrationTestDB(t)
		queries.ddl = nil
		if err := fresh.Table(name).AutoMigrate(value); err != nil {
			t.Fatal(err)
		}
		if len(queries.ddl) != 0 {
			t.Fatalf("run=%d repeated DDL=%v", i, queries.ddl)
		}
	}
}

type chContractProbe struct {
	ID        string    `gorm:"type:UUID"`
	Token     string    `gorm:"type:varchar(128);default:'';comment:会话令牌"`
	Count     int64     `gorm:"type:Int64;default:0"`
	Seen      time.Time `gorm:"type:DateTime64(3)"`
	Nullable  *float64  `gorm:"type:Nullable(Float64)"`
	NoDefault string    `gorm:"type:String"`
}

func TestClickHouseIdenticalSchemaConverges(t *testing.T) {
	db, q := clickHouseMigrationTestDB(t)
	name := chTable(t, db)
	if err := db.Table(name).AutoMigrate(&chContractProbe{}); err != nil {
		t.Fatal(err)
	}
	chExec(t, db, "INSERT INTO `"+name+"` (id,token,count,seen,nullable,no_default) VALUES ('00000000-0000-0000-0000-000000000001','中文Case',9223372036854775807,'2026-11-02 00:00:00.123',NULL,'retain')")
	chConverges(t, db, q, name, &chContractProbe{})
	cols, err := db.Table(name).Migrator().ColumnTypes(&chContractProbe{})
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range cols {
		_, valid := c.DefaultValue()
		if c.Name() == "token" && !valid {
			t.Fatal("空字符串默认值丢失")
		}
		if c.Name() == "no_default" && valid {
			t.Fatal("无默认值误报有默认值")
		}
	}
	var count int64
	var nullable uint8
	var token string
	if err := db.Raw("SELECT count,isNull(nullable),token FROM `"+name+"`").Row().Scan(&count, &nullable, &token); err != nil {
		t.Fatal(err)
	}
	if count != 9223372036854775807 || nullable != 1 || token != "中文Case" {
		t.Fatal("迁移改变了历史数据")
	}
}

func TestClickHouseAttributesMigrateWithoutType(t *testing.T) {
	type empty struct {
		Value string `gorm:"type:String;default:'';comment:新注释"`
	}
	type none struct {
		Value string `gorm:"type:String;comment:新注释"`
	}
	type quoted struct {
		Value string `gorm:"type:String;default:'Case''Quote';comment:新注释"`
	}
	for _, tc := range []struct {
		name, ddl string
		value     interface{}
		want      int
	}{
		{"add_empty", "String COMMENT '旧注释'", &empty{}, 2},
		{"remove_default", "String DEFAULT '' COMMENT '新注释'", &none{}, 1},
		{"replace_literal", "String DEFAULT 'case''Quote' COMMENT '新注释'", &quoted{}, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, q := clickHouseMigrationTestDB(t)
			name := chTable(t, db)
			chExec(t, db, "CREATE TABLE `"+name+"` (value "+tc.ddl+") ENGINE=MergeTree ORDER BY tuple()")
			chExec(t, db, "INSERT INTO `"+name+"` VALUES ('existing')")
			q.ddl = nil
			if err := db.Table(name).AutoMigrate(tc.value); err != nil {
				t.Fatal(err)
			}
			if len(q.ddl) != tc.want {
				t.Fatalf("DDL=%v", q.ddl)
			}
			for _, sql := range q.ddl {
				if strings.Contains(sql, " String") {
					t.Fatalf("属性迁移不应带类型: %s", sql)
				}
			}
			chConverges(t, db, q, name, tc.value)
			var s string
			if err := db.Raw("SELECT value FROM `" + name + "`").Scan(&s).Error; err != nil || s != "existing" {
				t.Fatalf("历史值改变: %s %v", s, err)
			}
		})
	}
}

func TestClickHouseRealTypesAndAddedColumn(t *testing.T) {
	for _, tc := range []struct{ name, old, tag string }{
		{"width", "FixedString(4)", "FixedString(8)"},
		{"decimal_scale", "Decimal(10,2)", "Decimal(10,3)"},
		{"decimal_precision", "Decimal(10,2)", "Decimal(12,2)"},
		{"datetime_precision", "DateTime64(3)", "DateTime64(6)"},
		{"nullable", "Int64", "Nullable(Int64)"},
		{"required", "Nullable(Int64)", "Int64"},
		{"signedness", "Int32", "UInt32"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, q := clickHouseMigrationTestDB(t)
			name := chTable(t, db)
			chExec(t, db, "CREATE TABLE `"+name+"` (value "+tc.old+") ENGINE=MergeTree ORDER BY tuple()")
			// 显式类型构造不同模型，确保长度/精度都参与真实迁移。
			value := chDynamicProbe(tc.tag)
			q.ddl = nil
			if err := db.Table(name).AutoMigrate(value); err != nil {
				t.Fatal(err)
			}
			if len(q.ddl) != 1 || !strings.Contains(q.ddl[0], tc.tag) {
				t.Fatalf("漏掉真实类型差异: %v", q.ddl)
			}
			chConverges(t, db, q, name, value)
		})
	}
	db, q := clickHouseMigrationTestDB(t)
	name := chTable(t, db)
	chExec(t, db, "CREATE TABLE `"+name+"` (id UInt64) ENGINE=MergeTree ORDER BY tuple()")
	type added struct {
		ID   uint64
		Note string `gorm:"type:String;default:'new';index:note_idx,type:minmax"`
	}
	q.ddl = nil
	if err := db.Table(name).AutoMigrate(&added{}); err != nil {
		t.Fatal(err)
	}
	if len(q.ddl) != 2 {
		t.Fatalf("补列/索引失败: %v", q.ddl)
	}
	chConverges(t, db, q, name, &added{})
}

func chDynamicProbe(tag string) interface{} {
	return reflect.New(reflect.StructOf([]reflect.StructField{{Name: "Value", Type: reflect.TypeOf(""), Tag: reflect.StructTag(`gorm:"type:` + tag + `"`)}})).Interface()
}

func TestClickHouseAliasComparisonPreservesRealDifferences(t *testing.T) {
	db, _ := clickHouseMigrationTestDB(t)
	m := db.Migrator().(clickHouseColumnMigrator)
	for _, tc := range []struct {
		a, b string
		same bool
	}{
		{"varchar(128)", "String", true},
		{"Nullable(varchar(16))", "Nullable(String)", true},
		{"Array(Nullable(varchar(16)))", "Array(Nullable(String))", true},
		{"Decimal32(2)", "Decimal(9,2)", true},
		{"FixedString(8)", "FixedString(16)", false},
		{"Nullable(Int64)", "Int64", false},
		{"Decimal(9,2)", "Decimal(9,3)", false},
		{"DateTime64(3, 'UTC')", "DateTime64(3, 'Asia/Shanghai')", false},
		{"Enum8('A'=1)", "Enum8('a'=1)", false},
	} {
		same, err := m.sameType(tc.a, tc.b)
		if err != nil || same != tc.same {
			t.Fatalf("%s / %s: same=%v error=%v", tc.a, tc.b, same, err)
		}
	}
	for _, tc := range []struct {
		a, b string
		same bool
	}{
		{"''", "", false}, {"'Case'", "'case'", false},
		{"1.000000000000000000001", "1.000000000000000000002", false},
		{"now ( )", "now()", true},
		{"throwIf(1)", "throwIf( 1 )", true},
	} {
		same, err := m.sameExpression(tc.a, tc.b, false)
		if err != nil || same != tc.same {
			t.Fatalf("%s / %s: same=%v error=%v", tc.a, tc.b, same, err)
		}
	}
}

func TestClickHouseDatabaseExpressionsAndDDLFailure(t *testing.T) {
	db, q := clickHouseMigrationTestDB(t)
	name := chTable(t, db)
	chExec(t, db, "CREATE TABLE `"+name+"` (base Int64, value Int64 MATERIALIZED base+1) ENGINE=MergeTree ORDER BY tuple()")
	type computed struct {
		Base  int64
		Value int64 `gorm:"default:(-)"`
	}
	chConverges(t, db, q, name, &computed{})
	type ordinary struct {
		Base  int64
		Value int64
	}
	q.ddl = nil
	if err := db.Table(name).AutoMigrate(&ordinary{}); err == nil || !strings.Contains(err.Error(), "MATERIALIZED") {
		t.Fatalf("未暴露数据库表达式的契约差异: %v", err)
	}
	if len(q.ddl) != 0 {
		t.Fatalf("未明确契约前不应修改表达式: %v", q.ddl)
	}
	fail := chTable(t, db)
	chExec(t, db, "CREATE TABLE `"+fail+"` (value String) ENGINE=MergeTree ORDER BY value")
	chExec(t, db, "INSERT INTO `"+fail+"` VALUES ('cannot_cast')")
	type numeric struct{ Value int64 }
	if err := db.Table(fail).AutoMigrate(&numeric{}); err == nil {
		t.Fatal("真实转换错误被忽略")
	}
	var s string
	if err := db.Raw("SELECT value FROM `" + fail + "`").Scan(&s).Error; err != nil || s != "cannot_cast" {
		t.Fatalf("失败改型破坏原数据: %s %v", s, err)
	}
}

func TestClickHouseTTLCodecPreservedAndExplicitlyChanged(t *testing.T) {
	db, q := clickHouseMigrationTestDB(t)
	name := chTable(t, db)
	chExec(t, db, "CREATE TABLE `"+name+"` (value DateTime DEFAULT now() COMMENT '旧' CODEC(ZSTD(1)) TTL value + INTERVAL 10 YEAR) ENGINE=MergeTree ORDER BY tuple()")
	type property struct {
		Value time.Time `gorm:"type:DateTime;default:now();comment:新"`
	}
	q.ddl = nil
	if err := db.Table(name).AutoMigrate(&property{}); err != nil {
		t.Fatal(err)
	}
	if len(q.ddl) != 1 || !strings.Contains(q.ddl[0], "COMMENT COLUMN") {
		t.Fatalf("注释差异附带改型: %v", q.ddl)
	}
	cols, err := db.Table(name).Migrator().ColumnTypes(&property{})
	if err != nil {
		t.Fatal(err)
	}
	state := cols[0].(clickHouseColumn).state
	if !strings.Contains(state.CodecExpression, "ZSTD(1)") || !strings.Contains(state.TtlExpression, "10") {
		t.Fatalf("丢失已有 codec/TTL: %+v", state)
	}
	type explicit struct {
		Value time.Time `gorm:"type:DateTime;default:now();comment:新;codec:ZSTD(3);ttl:value + INTERVAL 20 YEAR"`
	}
	q.ddl = nil
	if err := db.Table(name).AutoMigrate(&explicit{}); err != nil {
		t.Fatal(err)
	}
	if len(q.ddl) != 2 {
		t.Fatalf("漏掉 codec/TTL 差异: %v", q.ddl)
	}
	chConverges(t, db, q, name, &explicit{})
}

func TestClickHouseDryRunAndPrimaryKeyDifferences(t *testing.T) {
	db, q := clickHouseMigrationTestDB(t)
	name := chTable(t, db)
	chExec(t, db, "CREATE TABLE `"+name+"` (value Int32 DEFAULT 1 COMMENT '旧') ENGINE=MergeTree ORDER BY tuple()")
	type changed struct {
		Value int64 `gorm:"primaryKey;type:Int64;default:2;comment:新"`
	}
	q.ddl = nil
	if err := db.Session(&gorm.Session{DryRun: true}).Table(name).AutoMigrate(&changed{}); err != nil {
		t.Fatal(err)
	}
	if len(q.ddl) != 2 {
		t.Fatalf("DryRun 漏真实差异: %v", q.ddl)
	}
	var typ, def, comment string
	if err := db.Raw("SELECT type,default_expression,comment FROM system.columns WHERE database=currentDatabase() AND table=? AND name='value'", name).Row().Scan(&typ, &def, &comment); err != nil {
		t.Fatal(err)
	}
	if typ != "Int32" || def != "1" || comment != "旧" {
		t.Fatalf("DryRun 修改了结构: %s %s %s", typ, def, comment)
	}
	q.ddl = nil
	if err := db.Table(name).AutoMigrate(&changed{}); err != nil {
		t.Fatal(err)
	}
	if len(q.ddl) != 2 {
		t.Fatalf("主键字段真实差异被忽略: %v", q.ddl)
	}
	chConverges(t, db, q, name, &changed{})
}

// 分别只改变一个比较条件，沿未修复驱动路径确认重复 ALTER 的来源。
func TestClickHouseUpstreamRepeatedAlterCauses(t *testing.T) {
	for _, tc := range []struct {
		name, sql, tag, control string
		emptyFlag               bool
	}{
		{"alias", "String DEFAULT ''", "varchar(128);not null;default:''", "String;not null;default:''", false},
		{"nullability", "String DEFAULT ''", "String;default:''", "String;not null;default:''", false},
		{"absent_default", "Nullable(Float64)", "Nullable(Float64)", "Nullable(Float64)", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, _ := clickHouseMigrationTestDB(t)
			name := chTable(t, db)
			chExec(t, db, "CREATE TABLE `"+name+"` (value "+tc.sql+") ENGINE=MergeTree ORDER BY tuple()")
			dsn := os.Getenv("ANT_GODMIN_CLICKHOUSE_TEST_DSN")
			for _, control := range []bool{false, true} {
				l := &migrationStatements{Interface: logger.Default}
				upstream := clickhouse.Open(dsn).(*clickhouse.Dialector)
				upstream.DontSupportEmptyDefaultValue = control && tc.emptyFlag
				plain, err := gorm.Open(upstream, &gorm.Config{Logger: l})
				if err != nil {
					t.Fatal(err)
				}
				pool, _ := plain.DB()
				t.Cleanup(func() { _ = pool.Close() })
				tag := tc.tag
				if control {
					tag = tc.control
				}
				for i := 0; i < 2; i++ {
					l.ddl = nil
					if err := plain.Table(name).AutoMigrate(chDynamicProbe(tag)); err != nil {
						t.Fatal(err)
					}
					want := 1
					if control {
						want = 0
					}
					if len(l.ddl) != want {
						t.Fatalf("control=%v run=%d DDL=%v", control, i, l.ddl)
					}
				}
			}
		})
	}
}
