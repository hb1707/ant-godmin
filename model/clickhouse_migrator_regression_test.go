package model

import (
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/hb1707/ant-godmin/setting"
	"gorm.io/driver/clickhouse"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// 复审反例固定为回归测试，只允许在本机隔离库执行 DDL。
func TestClickHouseTypeAndDefaultTogether(t *testing.T) {
	for _, wrapped := range []bool{false, true} {
		t.Run(map[bool]string{false: "upstream", true: "adapter"}[wrapped], func(t *testing.T) {
			db, queries := clickHouseMigrationTestDB(t)
			if !wrapped {
				var err error
				db, err = gorm.Open(clickhouse.Open(os.Getenv("ANT_GODMIN_CLICKHOUSE_TEST_DSN")), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
				if err != nil {
					t.Fatal(err)
				}
				pool, _ := db.DB()
				t.Cleanup(func() { _ = pool.Close() })
			}
			name := chTable(t, db)
			chExec(t, db, "CREATE TABLE `"+name+"` (value String DEFAULT 'legacy') ENGINE=MergeTree ORDER BY tuple()")
			chExec(t, db, "INSERT INTO `"+name+"` VALUES ('42')")
			type desired struct {
				Value int64 `gorm:"type:Int64;default:0;not null"`
			}
			if err := db.Table(name).AutoMigrate(&desired{}); err != nil {
				var typ, def, stored string
				if e := db.Raw("SELECT type,default_expression FROM system.columns WHERE database=currentDatabase() AND table=? AND name='value'", name).Row().Scan(&typ, &def); e != nil {
					t.Fatal(e)
				}
				if e := db.Raw("SELECT value FROM `" + name + "`").Scan(&stored).Error; e != nil {
					t.Fatal(e)
				}
				t.Logf("失败后隔离表仍为 type=%s default=%s，历史值=%s", typ, def, stored)
				t.Fatalf("类型和默认值同时发生真实变化时应能完成有效迁移: %v", err)
			}
			var value int64
			if err := db.Raw("SELECT value FROM `" + name + "`").Scan(&value).Error; err != nil || value != 42 {
				t.Fatalf("历史值: %d %v", value, err)
			}
			if wrapped {
				chConverges(t, db, queries, name, &desired{})
			}
		})
	}
}

func TestClickHouseTypeAndCodecTogether(t *testing.T) {
	for _, wrapped := range []bool{false, true} {
		t.Run(map[bool]string{false: "upstream", true: "adapter"}[wrapped], func(t *testing.T) {
			db, queries := clickHouseMigrationTestDB(t)
			if !wrapped {
				var err error
				db, err = gorm.Open(clickhouse.Open(os.Getenv("ANT_GODMIN_CLICKHOUSE_TEST_DSN")), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
				if err != nil {
					t.Fatal(err)
				}
				pool, _ := db.DB()
				t.Cleanup(func() { _ = pool.Close() })
			}
			name := chTable(t, db)
			chExec(t, db, "CREATE TABLE `"+name+"` (value Int64 CODEC(Delta,ZSTD(1))) ENGINE=MergeTree ORDER BY tuple()")
			chExec(t, db, "INSERT INTO `"+name+"` VALUES (42)")
			type desired struct {
				Value string `gorm:"type:String;codec:ZSTD(1);default:(-);not null"`
			}
			if err := db.Table(name).AutoMigrate(&desired{}); err != nil {
				t.Fatalf("类型和codec同时发生真实变化时应能完成有效迁移: %v", err)
			}
			var value string
			if err := db.Raw("SELECT value FROM `" + name + "`").Scan(&value).Error; err != nil || value != "42" {
				t.Fatalf("历史值: %s %v", value, err)
			}
			if wrapped {
				chConverges(t, db, queries, name, &desired{})
			}
		})
	}
}

func TestClickHouseHexDefaultConverges(t *testing.T) {
	db, _ := clickHouseMigrationTestDB(t)
	name := chTable(t, db)
	type desired struct {
		Value uint64 `gorm:"type:UInt64;default:bitAnd(255,0xFF)"`
	}
	if err := db.Table(name).AutoMigrate(&desired{}); err != nil {
		t.Fatal(err)
	}

	for run := 0; run < 3; run++ {
		fresh, trace := clickHouseMigrationTestDB(t)
		if err := fresh.Table(name).AutoMigrate(&desired{}); err != nil {
			t.Fatal(err)
		}
		if len(trace.ddl) != 0 {
			t.Errorf("新连接第%d轮仍执行重复DDL: %v", run+1, trace.ddl)
		}
	}
}

func TestClickHouseIntegerTypeComparison(t *testing.T) {
	db, _ := clickHouseMigrationTestDB(t)
	m := db.Migrator().(clickHouseColumnMigrator)
	for _, pair := range [][2]string{{"Tuple(a Int64)", "Tuple(b Int64)"}, {"Array(Int64)", "Array(UInt64)"}, {"String", "LowCardinality(String)"}, {"DateTime", "DateTime64(3)"}} {
		same, err := m.sameType(pair[0], pair[1])
		if err != nil || same {
			t.Fatalf("真实类型差异被遗漏: %v same=%v err=%v", pair, same, err)
		}
	}
}

func TestClickHousePublicEntryPointFlags(t *testing.T) {
	db, _ := clickHouseMigrationTestDB(t)
	name := chTable(t, db)
	chExec(t, db, "CREATE TABLE `"+name+"` (value String) ENGINE=MergeTree ORDER BY tuple()")
	oldDB, oldConf := CHDB, setting.ClickHouse
	t.Cleanup(func() { CHDB = oldDB; setting.ClickHouse = oldConf })
	u, err := url.Parse(os.Getenv("ANT_GODMIN_CLICKHOUSE_TEST_DSN"))
	if err != nil {
		t.Fatal(err)
	}
	setting.ClickHouse.ENABLE = true
	setting.ClickHouse.HOST = u.Hostname()
	setting.ClickHouse.PORT = u.Port()
	setting.ClickHouse.DATABASE = "center_orm_repro"
	setting.ClickHouse.USERNAME = "default"
	setting.ClickHouse.PASSWORD = ""
	setting.ClickHouse.OPTIONS = "?dial_timeout=10s&read_timeout=20s"
	type desired struct {
		Value string `gorm:"type:String;default:''"`
	}
	for _, enabled := range []bool{false, true, true} {
		setting.ClickHouse.AUTOMIGRATE = enabled
		if err := OpenClickHouse(); err != nil {
			t.Fatal(err)
		}
		if _, ok := CHDB.Migrator().(clickHouseColumnMigrator); !ok {
			t.Fatal("公开入口没有使用专用适配")
		}
		pool, _ := CHDB.DB()
		t.Cleanup(func() { _ = pool.Close() })
		q := &migrationStatements{Interface: logger.Default}
		CHDB = CHDB.Session(&gorm.Session{Logger: q}).Table(name)
		cols, err := CHDB.Migrator().ColumnTypes(&desired{})
		if err != nil {
			t.Fatal(err)
		}
		_, hasDefault := cols[0].DefaultValue()
		if err := CreateTableClickHouse(&desired{}); err != nil {
			t.Fatal(err)
		}
		want := 0
		if enabled && !hasDefault {
			want = 1
		}
		if len(q.ddl) != want {
			t.Fatalf("CH_AUTO_MIGRATE=%v DDL=%v", enabled, q.ddl)
		}
	}
	setting.ClickHouse.ENABLE = false
	if err := OpenClickHouse(); err != nil || CHDB != nil {
		t.Fatal("禁用 CH 后不应保留活动连接")
	}
}

func TestClickHouseJointColumnChangesConverge(t *testing.T) {
	for _, tc := range []struct {
		name, old, insert, tag, wantType, wantDefault, wantValue string
		ddl                                                      int
	}{
		{"replace_default", "String DEFAULT 'legacy'", "'42'", "type:Int64;default:0", "Int64", "0", "42", 1},
		{"new_default_invalid_for_old_type", "UInt64 DEFAULT 42", "42", "type:String;default:'new'", "String", "'new'", "42", 1},
		{"remove_default", "String DEFAULT 'legacy'", "'42'", "type:Int64", "Int64", "", "42", 2},
		{"remove_nullable_and_null_default", "Nullable(Int64) DEFAULT NULL", "42", "type:Int64;default:0", "Int64", "0", "42", 1},
		{"add_ttl_and_change_codec", "String DEFAULT 'legacy' CODEC(ZSTD(1))", "'2030-01-01 00:00:00.000'", "type:DateTime64(3,'UTC');default:now64(3);codec:ZSTD(3);ttl:toDateTime(value) + INTERVAL 20 YEAR", "DateTime64(3, 'UTC')", "now64(3)", "2030-01-01 00:00:00.000", 1},
		{"remove_ttl_and_change_codec", "DateTime DEFAULT now() CODEC(Delta,ZSTD(1)) TTL value + INTERVAL 10 YEAR", "toDateTime('2030-01-01 00:00:00','UTC')", "type:Int64;default:0;codec:ZSTD(3);ttl:", "Int64", "0", "1893456000", 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, q := clickHouseMigrationTestDB(t)
			name := chTable(t, db)
			chExec(t, db, "CREATE TABLE `"+name+"` (value "+tc.old+") ENGINE=MergeTree ORDER BY tuple()")
			chExec(t, db, "INSERT INTO `"+name+"` VALUES ("+tc.insert+")")
			// tag 控制真实字段契约，Go 类型不承担表达式归一化。
			goType := reflect.TypeOf(int64(0))
			if tc.wantType == "String" {
				goType = reflect.TypeOf("")
			}
			if strings.HasPrefix(tc.wantType, "DateTime") {
				goType = reflect.TypeOf(time.Time{})
			}
			value := reflect.New(reflect.StructOf([]reflect.StructField{{Name: "Value", Type: goType, Tag: reflect.StructTag(`gorm:"` + tc.tag + `"`)}})).Interface()
			q.ddl = nil
			if err := db.Table(name).AutoMigrate(value); err != nil {
				t.Fatal(err)
			}
			if len(q.ddl) != tc.ddl {
				t.Fatalf("DDL=%v", q.ddl)
			}
			var typ, def, stored string
			if err := db.Raw("SELECT type,default_expression FROM system.columns WHERE database=currentDatabase() AND table=? AND name='value'", name).Row().Scan(&typ, &def); err != nil {
				t.Fatal(err)
			}
			if err := db.Raw("SELECT toString(value) FROM `" + name + "`").Row().Scan(&stored); err != nil {
				t.Fatal(err)
			}
			if typ != tc.wantType || def != tc.wantDefault || stored != tc.wantValue {
				t.Fatalf("type=%s default=%s value=%s", typ, def, stored)
			}
			if strings.Contains(tc.tag, "ttl:") {
				columns, err := db.Table(name).Migrator().ColumnTypes(value)
				if err != nil {
					t.Fatal(err)
				}
				state := columns[0].(clickHouseColumn).state
				wantTTL := !strings.HasSuffix(tc.tag, "ttl:")
				if (state.TtlExpression != "") != wantTTL || state.CodecExpression != "ZSTD(3)" {
					t.Fatalf("TTL=%s codec=%s", state.TtlExpression, state.CodecExpression)
				}
			}
			chConverges(t, db, q, name, value)
		})
	}
}

func TestClickHouseNumericLiteralsPreserveDifferences(t *testing.T) {
	db, _ := clickHouseMigrationTestDB(t)
	m := db.Migrator().(clickHouseColumnMigrator)
	for _, tc := range []struct {
		a, b string
		same bool
	}{
		{"bitAnd(255,0xFF)", "bitAnd(255,255)", true},
		{"bitAnd(255,0b1111_1111)", "bitAnd(255,255)", true},
		{"tuple(010,1_000,1.5e1)", "tuple(10,1000,15.0)", true},
		{"tuple(0x1.fp3,0xFF_FF)", "tuple(15.5,65535)", true},
		{"bitAnd(255,0xFE)", "bitAnd(255,255)", false},
		{"bitAnd(255,0b11111110)", "bitAnd(255,255)", false},
		{"1.000000000000000000001", "1.000000000000000000002", false},
		{"0xFFFFFFFFFFFFFFFE", "0xFFFFFFFFFFFFFFFF", false},
		{"tuple('0xFF',1)", "tuple('255',1)", false},
	} {
		same, err := m.sameExpression(tc.a, tc.b, false)
		if err != nil || same != tc.same {
			t.Errorf("%s / %s: same=%v err=%v", tc.a, tc.b, same, err)
		}
	}
	name := chTable(t, db)
	chExec(t, db, "CREATE TABLE `"+name+"` (value UInt64 DEFAULT bitAnd(255,255), marker UInt8) ENGINE=MergeTree ORDER BY tuple()")
	type changed struct {
		Value uint64 `gorm:"type:UInt64;default:bitAnd(255,0xFE)"`
	}
	fresh, q := clickHouseMigrationTestDB(t)
	if err := fresh.Table(name).AutoMigrate(&changed{}); err != nil {
		t.Fatal(err)
	}
	if len(q.ddl) != 1 {
		t.Fatalf("漏掉真实默认值差异: %v", q.ddl)
	}
	chExec(t, db, "INSERT INTO `"+name+"` (marker) VALUES (1)")
	var value uint64
	if err := db.Raw("SELECT value FROM `" + name + "`").Row().Scan(&value); err != nil || value != 254 {
		t.Fatalf("value=%d err=%v", value, err)
	}
	chConverges(t, fresh, q, name, &changed{})
}

func TestClickHouseJointChangeErrorsRemainVisible(t *testing.T) {
	for _, removeDefault := range []bool{false, true} {
		t.Run(map[bool]string{false: "invalid_new_default", true: "rejected_key_type"}[removeDefault], func(t *testing.T) {
			db, _ := clickHouseMigrationTestDB(t)
			name := chTable(t, db)
			order := "tuple()"
			tag := "type:Int64;default:toInt64('invalid')"
			if removeDefault {
				order = "value"
				tag = "type:Int64"
			}
			chExec(t, db, "CREATE TABLE `"+name+"` (value String DEFAULT 'legacy') ENGINE=MergeTree ORDER BY "+order)
			chExec(t, db, "INSERT INTO `"+name+"` VALUES ('retain')")
			value := reflect.New(reflect.StructOf([]reflect.StructField{{Name: "Value", Type: reflect.TypeOf(int64(0)), Tag: reflect.StructTag(`gorm:"` + tag + `"`)}})).Interface()
			for run := 0; run < 2; run++ {
				fresh, _ := clickHouseMigrationTestDB(t)
				if err := fresh.Table(name).AutoMigrate(value); err == nil {
					t.Fatal("真实转换错误被忽略")
				}
			}
			var typ, def, stored string
			if err := db.Raw("SELECT type,default_expression FROM system.columns WHERE database=currentDatabase() AND table=? AND name='value'", name).Row().Scan(&typ, &def); err != nil {
				t.Fatal(err)
			}
			if err := db.Raw("SELECT value FROM `" + name + "`").Row().Scan(&stored); err != nil {
				t.Fatal(err)
			}
			wantDefault := "'legacy'"
			// ClickHouse DDL 不支持事务回滚：独立移除默认值可能成功，随后改型失败。
			// 保留已完成的真实差异并继续暴露失败，不能吞错或伪报全部迁移完成。
			if removeDefault {
				wantDefault = ""
			}
			if typ != "String" || def != wantDefault || stored != "retain" {
				t.Fatalf("type=%s default=%s value=%s", typ, def, stored)
			}
		})
	}
}
