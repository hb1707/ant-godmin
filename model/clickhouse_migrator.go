package model

import (
	"fmt"
	"math/big"
	"regexp"
	"strings"

	"gorm.io/driver/clickhouse"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

// ClickHouse 的可空性是类型的一部分，默认值也有独立的存在位；不能套用
// 通用 SQL 的 NotNull 和空字符串比较，否则每次启动都会重复 MODIFY COLUMN。
type clickHouseDialector struct{ *clickhouse.Dialector }

func (d clickHouseDialector) Migrator(db *gorm.DB) gorm.Migrator {
	upstream := d.Dialector.Migrator(db).(clickhouse.Migrator)
	upstream.Migrator.Config.Dialector = d
	return clickHouseColumnMigrator{Migrator: upstream}
}

type clickHouseColumnMigrator struct{ clickhouse.Migrator }
type clickHouseColumnState struct {
	Name              string
	Type              string
	DefaultType       string
	DefaultExpression string
	Comment           string
	CodecExpression   string
	TtlExpression     string
}
type clickHouseColumnType = gorm.ColumnType

type clickHouseColumn struct {
	clickHouseColumnType
	state clickHouseColumnState
}

func (c clickHouseColumn) DefaultValue() (string, bool) {
	if c.state.DefaultType != "DEFAULT" {
		return "", false
	}
	// 保留上游对字面量的返回形式，同时纠正无默认值的存在位。
	value, _ := c.clickHouseColumnType.DefaultValue()
	return value, true
}
func (m clickHouseColumnMigrator) ColumnTypes(value interface{}) ([]gorm.ColumnType, error) {
	columns, err := m.Migrator.ColumnTypes(value)
	if err != nil {
		return nil, err
	}
	err = m.RunWithValue(value, func(stmt *gorm.Statement) error {
		var states []clickHouseColumnState
		if err := m.DB.Raw("DESCRIBE TABLE ?", clause.Table{Name: stmt.Table}).Scan(&states).Error; err != nil {
			return err
		}
		byName := make(map[string]clickHouseColumnState, len(states))
		for _, s := range states {
			byName[s.Name] = s
		}
		for i, c := range columns {
			s, ok := byName[c.Name()]
			if !ok {
				return fmt.Errorf("ClickHouse column metadata missing: %s.%s", stmt.Table, c.Name())
			}
			columns[i] = clickHouseColumn{clickHouseColumnType: c, state: s}
		}
		return nil
	})
	return columns, err
}

// DryRun 仍读取真实元数据，但不能执行 DDL；Session(false) 本身不会清除继承的 DryRun。
func (m clickHouseColumnMigrator) metadataDB() *gorm.DB {
	db := m.DB.Session(&gorm.Session{})
	db.DryRun = false
	return db
}

// 借助服务器解析空数组的元素类型来展开别名，无业务行转换，也不执行默认表达式。
// 这样保留 FixedString 长度、Decimal 精度、Nullable、时区、Enum 值及嵌套类型的真实差异。
func (m clickHouseColumnMigrator) sameType(a, b string) (bool, error) {
	if a == b {
		return true, nil
	}
	var left, right string
	if err := m.metadataDB().Raw("SELECT toTypeName(CAST([], ?)), toTypeName(CAST([], ?))", "Array("+a+")", "Array("+b+")").Row().Scan(&left, &right); err != nil {
		return false, err
	}
	return left == right, nil
}

// formatQuerySingleLine 只解析 SQL，不求值。大小写敏感的字符串、枚举值及
// 高精度数值不能通过 ToLower、去引号或浮点近似进行比较。
func (m clickHouseColumnMigrator) sameExpression(a, b string, codec bool) (bool, error) {
	if a == b {
		return true, nil
	}
	if a == "" || b == "" {
		return false, nil
	}
	// ClickHouse 的 AST formatter 会把未带显式 Decimal 转换的小数解析为
	// Float64；先精确核对数值 token，防止高精度真实差异在格式化时被舍入掉。
	if clickHouseNumericSignature(a) != clickHouseNumericSignature(b) {
		return false, nil
	}
	leftSQL, rightSQL := "SELECT "+a, "SELECT "+b
	if codec {
		leftSQL = "CREATE TABLE _ch_contract (c UInt8 " + a + ") ENGINE=Memory"
		rightSQL = "CREATE TABLE _ch_contract (c UInt8 " + b + ") ENGINE=Memory"
	}
	var left, right string
	if err := m.metadataDB().Raw("SELECT formatQuerySingleLine(?), formatQuerySingleLine(?)", leftSQL, rightSQL).Row().Scan(&left, &right); err != nil {
		return false, err
	}
	return left == right, nil
}

func (m clickHouseColumnMigrator) MigrateColumn(value interface{}, field *schema.Field, column gorm.ColumnType) error {
	if field.IgnoreMigration {
		return nil
	}
	current, ok := column.(clickHouseColumn)
	if !ok {
		return fmt.Errorf("ClickHouse migration requires DESCRIBE metadata for %s", field.DBName)
	}
	return m.RunWithValue(value, func(stmt *gorm.Statement) error {
		state := current.state
		typeField := *field
		desiredType := m.Migrator.Migrator.DataTypeOf(&typeField)
		sameType, err := m.sameType(desiredType, state.Type)
		if err != nil {
			return err
		}
		defaultSQL := ""
		if field.HasDefaultValue && field.DefaultValue != "(-)" {
			if field.DefaultValueInterface != nil {
				defaultSQL = m.Dialector.Explain("?", field.DefaultValueInterface)
			} else {
				defaultSQL = field.DefaultValue
			}
		}
		// MATERIALIZED/ALIAS 的所有权不能从普通 DEFAULT tag 猜测；必须显式迁移，
		// 或用 default:(-) 声明由数据库维护，不能悄悄抹掉其表达式。
		if field.DefaultValue != "(-)" && state.DefaultType != "" && state.DefaultType != "DEFAULT" {
			return fmt.Errorf("ClickHouse %s.%s has %s expression; explicit migration or default:(-) required", stmt.Table, field.DBName, state.DefaultType)
		}
		sameDefault := true
		if field.DefaultValue != "(-)" {
			sameDefault, err = m.sameExpression(defaultSQL, state.DefaultExpression, false)
			if err != nil {
				return err
			}
			if (defaultSQL != "") != (state.DefaultType == "DEFAULT") {
				sameDefault = false
			}
		}
		sameTTL, sameCodec := true, true
		ttl, hasTTL := field.TagSettings["TTL"]
		if hasTTL {
			sameTTL, err = m.sameExpression(ttl, state.TtlExpression, false)
			if err != nil {
				return err
			}
		}
		codec, hasCodec := field.TagSettings["CODEC"]
		if hasCodec && codec != "" {
			codec = "CODEC(" + codec + ")"
			currentCodec := state.CodecExpression
			if currentCodec != "" {
				currentCodec = "CODEC(" + currentCodec + ")"
			}
			sameCodec, err = m.sameExpression(codec, currentCodec, true)
			if err != nil {
				return err
			}
		}
		cluster := ""
		if opt, ok := m.DB.Get("gorm:table_cluster_options"); ok {
			cluster = " " + fmt.Sprint(opt)
		}
		modify := func(sql string, args ...interface{}) error {
			params := []interface{}{clause.Table{Name: stmt.Table}, clause.Column{Name: field.DBName}}
			return m.DB.Exec("ALTER TABLE ?"+cluster+" MODIFY COLUMN ? "+sql, append(params, args...)...).Error
		}
		// 默认值与注释分别修改，既不附带改型，也不丢弃已有的 codec/TTL/默认值。
		if !sameType {
			if err := modify("?", clause.Expr{SQL: desiredType}); err != nil {
				return err
			}
		}
		if !sameDefault {
			if defaultSQL == "" {
				err = modify("REMOVE DEFAULT")
			} else {
				err = modify("DEFAULT ?", clause.Expr{SQL: defaultSQL})
			}
			if err != nil {
				return err
			}
		}
		if state.Comment != field.Comment {
			if err := m.DB.Exec("ALTER TABLE ?"+cluster+" COMMENT COLUMN ? ?", clause.Table{Name: stmt.Table}, clause.Column{Name: field.DBName}, field.Comment).Error; err != nil {
				return err
			}
		}
		if !sameTTL {
			if strings.TrimSpace(ttl) == "" {
				err = modify("REMOVE TTL")
			} else {
				// ClickHouse 24.3 修改列 TTL 的语法要求包含类型；此时 TTL 本身存在真实差异。
				err = modify("? TTL ?", clause.Expr{SQL: desiredType}, clause.Expr{SQL: ttl})
			}
			if err != nil {
				return err
			}
		}
		if !sameCodec {
			if err := modify("?", clause.Expr{SQL: codec}); err != nil {
				return err
			}
		}
		return nil
	})
}

// SQL 引号内的数字是字符串，不参与数值归一；标识符中的位宽也不作为独立数字。
var clickHouseExpressionTokens = regexp.MustCompile("(?s:/\\*.*?\\*/)|--[^\\n]*|'(?:[^'\\\\]|\\\\.|'')*'|\"(?:[^\"\\\\]|\\\\.|\"\")*\"|`[^`]*`|[a-zA-Z_][a-zA-Z_0-9]*|(?:[0-9]+(?:\\.[0-9]*)?|\\.[0-9]+)(?:[eE][+-]?[0-9]+)?")

func clickHouseNumericSignature(sql string) string {
	var numbers []string
	for _, token := range clickHouseExpressionTokens.FindAllString(sql, -1) {
		if (token[0] >= '0' && token[0] <= '9') || token[0] == '.' {
			if n, ok := new(big.Rat).SetString(token); ok {
				numbers = append(numbers, n.RatString())
			} else {
				numbers = append(numbers, token)
			}
		}
	}
	return strings.Join(numbers, ";")
}
