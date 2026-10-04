package model

import (
	"database/sql"
	"sync"
	"testing"

	"gorm.io/gorm/migrator"
	"gorm.io/gorm/schema"
)

type alignProbe struct {
	ID        uint    `gorm:"primaryKey;comment:主键"`
	NoComment string  `gorm:"type:varchar(24);not null;default:'project'"`
	Declared  string  `gorm:"type:varchar(24);comment:模型注释"`
	Meta      []byte  `gorm:"type:jsonb;not null;default:'{}'"`
	List      []byte  `gorm:"type:jsonb;not null;default:'[]'"`
	Name      string  `gorm:"type:varchar(64);not null;default:''"`
	Score     float64 `gorm:"type:decimal(5,4);default:1.0"`
	Weight    float64 `gorm:"type:decimal(5,4);default:0.5"`
}

func existingColumn(name, comment, defaultValue string) migrator.ColumnType {
	return migrator.ColumnType{
		NameValue:         sql.NullString{String: name, Valid: true},
		CommentValue:      sql.NullString{String: comment, Valid: comment != ""},
		DefaultValueValue: sql.NullString{String: defaultValue, Valid: defaultValue != ""},
	}
}

// 只消除「改了也不会收敛」的两类假差异；模型显式声明了不同的值时照常迁移。
func TestAlignFieldWithExistingColumn(t *testing.T) {
	parsed, err := schema.Parse(&alignProbe{}, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		t.Fatal(err)
	}
	field := func(name string) *schema.Field { return parsed.LookUpField(name) }

	// 模型没写注释、库里有：以库为准，否则每次启动都 ALTER COLUMN TYPE。
	alignFieldWithExistingColumn(field("no_comment"), existingColumn("no_comment", "迁移写入的注释", "project"))
	if got := field("no_comment").Comment; got != "迁移写入的注释" {
		t.Errorf("未声明注释应采用库中注释, got %q", got)
	}
	// 模型显式声明了注释：不覆盖，差异照常交给 AutoMigrate。
	alignFieldWithExistingColumn(field("declared"), existingColumn("declared", "库里旧注释", ""))
	if got := field("declared").Comment; got != "模型注释" {
		t.Errorf("显式注释不应被覆盖, got %q", got)
	}
	// jsonb 默认值：库里读回 {}，模型写 '{}'，等价，应对齐。
	alignFieldWithExistingColumn(field("meta"), existingColumn("meta", "", "{}"))
	if got := field("meta").DefaultValue; got != "{}" {
		t.Errorf("等价的 jsonb 默认值应对齐, got %q", got)
	}
	// 真实不同的默认值：不干预。
	alignFieldWithExistingColumn(field("list"), existingColumn("list", "", "{}"))
	if got := field("list").DefaultValue; got != "'[]'" {
		t.Errorf("不同的默认值不应被改写, got %q", got)
	}
	// 字符串类型 gorm 已有等价比较，不干预。
	before := field("name").DefaultValue
	alignFieldWithExistingColumn(field("name"), existingColumn("name", "", "x"))
	if got := field("name").DefaultValue; got != before {
		t.Errorf("字符串默认值不应被改写, got %q want %q", got, before)
	}
	// 数值默认值按值比较：库里 1、模型 1.0 视为相同；真实不同的数值不干预。
	alignFieldWithExistingColumn(field("score"), existingColumn("score", "", "1"))
	if got := field("score").DefaultValue; got != "1" {
		t.Errorf("相等的数值默认值应对齐, got %q", got)
	}
	alignFieldWithExistingColumn(field("weight"), existingColumn("weight", "", "1"))
	if got := field("weight").DefaultValue; got != "0.5" {
		t.Errorf("不同的数值默认值不应被改写, got %q", got)
	}
	// 主键不处理。
	alignFieldWithExistingColumn(field("id"), existingColumn("id", "库里主键注释", ""))
	if got := field("id").Comment; got != "主键" {
		t.Errorf("主键不应被处理, got %q", got)
	}
}
