package model

import (
	"database/sql"
	"math/big"
	"regexp"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

type postgresDialector struct {
	*postgres.Dialector
}

func newPostgresDialector(config postgres.Config) gorm.Dialector {
	return postgresDialector{postgres.New(config).(*postgres.Dialector)}
}

func (d postgresDialector) Migrator(db *gorm.DB) gorm.Migrator {
	queryMigrator := d.Dialector.Migrator(db).(postgres.Migrator)
	upstream := queryMigrator
	upstream.Config.Dialector = d
	return postgresColumnMigrator{Migrator: upstream, queryMigrator: queryMigrator}
}

func (d postgresDialector) DataTypeOf(field *schema.Field) string {
	declared := d.Dialector.DataTypeOf(field)
	// 统一同义类型的拼写，但完整保留长度/精度。否则上游的别名前缀比较会
	// 把 decimal(8,1) 与 decimal(10,2)、character varying(64) 与 (128) 判为相同。
	for _, alias := range [][2]string{{"character varying", "varchar"}, {"decimal", "numeric"}, {"character", "bpchar"}, {"char", "bpchar"}} {
		lower := strings.ToLower(declared)
		if lower == alias[0] || strings.HasPrefix(lower, alias[0]+"(") {
			return alias[1] + declared[len(alias[0]):]
		}
	}
	return declared
}

type postgresColumnMigrator struct {
	postgres.Migrator
	queryMigrator postgres.Migrator
}

func (m postgresColumnMigrator) ColumnTypes(value interface{}) ([]gorm.ColumnType, error) {
	// 上游 GetRows 通过具体方言类型识别 pgx simple protocol。读取元数据必须保留
	// 原方言，否则改型后的 SELECT * 会复用旧结果类型的执行计划而启动失败。
	return m.queryMigrator.ColumnTypes(value)
}

// 类型比较阶段不参与默认值比较；这些差异在下面单独迁移，不会被忽略。
type postgresTypeColumn struct {
	postgresColumnType
}

type postgresColumnType = gorm.ColumnType

func (postgresTypeColumn) DefaultValue() (string, bool) { return "", false }

type postgresColumnState struct {
	notNull    bool
	defaultSQL sql.NullString
	typeSQL    string
	nativeType string
	literalSQL sql.NullString
	typeMod    int32
}

// MigrateColumn 保留上游对类型、长度、精度、序列和唯一约束的处理，
// 将默认值、非空及注释差异拆开，避免它们触发驱动附带的 ALTER COLUMN TYPE。
func (m postgresColumnMigrator) MigrateColumn(value interface{}, field *schema.Field, column gorm.ColumnType) error {
	if field.IgnoreMigration {
		return nil
	}
	if !field.PrimaryKey {
		// 只复制字段，不改共享 schema 缓存；类型比较阶段仅使用数据库当前的非类型属性。
		typeField := *field
		typeField.HasDefaultValue = false
		typeField.DefaultValue = ""
		typeField.DefaultValueInterface = nil
		typeField.Comment, _ = column.Comment()
		if existingSequence, _ := column.AutoIncrement(); field.AutoIncrement && existingSequence {
			// serial 是创建语法，真实列类型仍为 integer/bigint；只比较底层类型，
			// 真正需要改型时上游仍通过原始模型维护序列。
			typeField.AutoIncrement = false
		}
		if nullable, ok := column.Nullable(); ok {
			typeField.NotNull = !nullable
		}
		// 无精度 numeric 与 numeric(p,s) 是真实差异。上游只在模型声明 p/s 时比较，
		// 反向解除数据库已有的精度限制也必须执行真实改型，不能把两者判为相同。
		if strings.EqualFold(strings.TrimSpace(m.DataTypeOf(&typeField)), "numeric") {
			if err := m.RunWithValue(value, func(stmt *gorm.Statement) error {
				current, err := m.postgresColumnState(stmt, field)
				if err != nil {
					return err
				}
				// pgx 对无限精度 numeric 也可能报告哨兵 precision，必须以 catalog typmod 为准。
				if current.nativeType != "numeric" || current.typeMod < 0 {
					return nil
				}
				col := clause.Column{Name: field.DBName}
				return m.DB.Exec("ALTER TABLE ? ALTER COLUMN ? TYPE numeric USING ?::numeric", m.CurrentTable(stmt), col, col).Error
			}); err != nil {
				return err
			}
		}

		// 使用 GORM 核心的类型/唯一约束比较，不调用 PostgreSQL 外层的注释更新。
		if err := m.Migrator.Migrator.MigrateColumn(value, &typeField, postgresTypeColumn{column}); err != nil {
			return err
		}
	}
	return m.RunWithValue(value, func(stmt *gorm.Statement) error {
		col := clause.Column{Name: field.DBName}
		if !field.PrimaryKey {
			// 类型变更可能已调整默认值或非空属性，因此读取当前状态，避免重复 DDL。
			// 必须使用原始默认表达式：驱动的 DefaultValue 会截断字符串里的 ::，且不还原转义。
			current, err := m.postgresColumnState(stmt, field)
			if err != nil {
				return err
			}
			// PostgreSQL 的 serial/identity 创建语义自带 NOT NULL。
			wantedNotNull := field.NotNull || field.AutoIncrement
			if current.notNull != wantedNotNull {
				operation := "DROP NOT NULL"
				if wantedNotNull {
					operation = "SET NOT NULL"
				}
				if err := m.DB.Exec("ALTER TABLE ? ALTER COLUMN ? "+operation, m.CurrentTable(stmt), col).Error; err != nil {
					return err
				}
			}
			// 无显式默认表达式的自增字段由上游序列迁移维护，不能写入空的 SET DEFAULT。
			managedBySequence := field.AutoIncrement && field.DefaultValueInterface == nil && field.DefaultValue == ""
			if field.DefaultValue != "(-)" && !managedBySequence {
				defaultSQL := field.TagSettings["DEFAULT"]
				if defaultSQL == "" {
					defaultSQL = field.DefaultValue
				}
				if _, literal := field.DefaultValueInterface.(string); literal {
					defaultSQL = current.literalSQL.String
				} else if field.DefaultValueInterface != nil {
					defaultSQL = m.Dialector.Explain("$1", field.DefaultValueInterface)
				}
				if !postgresDefaultMatches(field, current, defaultSQL) {
					if !field.HasDefaultValue || (field.DefaultValueInterface == nil && strings.EqualFold(field.DefaultValue, "NULL")) {
						if err := m.DB.Exec("ALTER TABLE ? ALTER COLUMN ? DROP DEFAULT", m.CurrentTable(stmt), col).Error; err != nil {
							return err
						}
					} else if err := m.DB.Exec("ALTER TABLE ? ALTER COLUMN ? SET DEFAULT ?", m.CurrentTable(stmt), col, clause.Expr{SQL: defaultSQL}).Error; err != nil {
						return err
					}
				}
			}
		}
		if comment, _ := column.Comment(); field.Comment != "" && field.Comment != comment {
			return m.DB.Exec("COMMENT ON COLUMN ?.? IS ?", m.CurrentTable(stmt), col,
				clause.Expr{SQL: m.Dialector.Explain("$1", field.Comment)}).Error
		}
		return nil
	})
}

func (m postgresColumnMigrator) postgresColumnState(stmt *gorm.Statement, field *schema.Field) (postgresColumnState, error) {
	var current postgresColumnState
	var literal interface{}
	if text, ok := field.DefaultValueInterface.(string); ok {
		literal = text
	}
	currentSchema, table := m.CurrentSchema(stmt, stmt.Table)
	queryDB := m.DB
	if queryDB.DryRun {
		// 与上游 queryRaw 保持一致：DryRun 仍需真实读取元数据，只禁止执行 DDL。
		queryDB = queryDB.Session(&gorm.Session{})
		queryDB.DryRun = false
	}
	// 仅反解析 catalog，不执行默认表达式，避免 uuid/sequence 等函数产生副作用。
	// quote_literal 交给 PostgreSQL 处理引号和反斜杠，不能靠字符串 Trim 还原 SQL 字面量。
	err := queryDB.Raw(`SELECT a.attnotnull, pg_catalog.pg_get_expr(d.adbin, d.adrelid),
		pg_catalog.format_type(a.atttypid, NULL),
		CASE WHEN t.typnamespace = 'pg_catalog'::regnamespace THEN t.typname ELSE pg_catalog.format_type(a.atttypid, NULL) END,
		pg_catalog.quote_literal(CAST(? AS text)), a.atttypmod
		FROM pg_catalog.pg_attribute a
		JOIN pg_catalog.pg_type t ON t.oid = a.atttypid
		JOIN pg_catalog.pg_class c ON c.oid = a.attrelid
		JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
		LEFT JOIN pg_catalog.pg_attrdef d ON d.adrelid = a.attrelid AND d.adnum = a.attnum
		WHERE n.nspname = ? AND c.relname = ? AND a.attname = ? AND a.attnum > 0 AND NOT a.attisdropped`,
		literal, currentSchema, table, field.DBName).Row().Scan(&current.notNull, &current.defaultSQL, &current.typeSQL, &current.nativeType, &current.literalSQL, &current.typeMod)
	return current, err
}

func postgresDefaultMatches(field *schema.Field, current postgresColumnState, wantedSQL string) bool {
	wanted := field.HasDefaultValue && (field.DefaultValueInterface != nil || !strings.EqualFold(field.DefaultValue, "NULL"))
	if wanted != current.defaultSQL.Valid {
		return false
	}
	if !wanted {
		return true
	}
	actual := current.defaultSQL.String
	if actual == wantedSQL || actual == wantedSQL+"::"+current.typeSQL || (current.nativeType != "" && actual == wantedSQL+"::"+current.nativeType) {
		return true
	}
	if _, literal := field.DefaultValueInterface.(string); literal {
		return false
	}
	switch field.GORMDataType {
	case schema.Int, schema.Uint, schema.Float:
		// PostgreSQL 可把整数默认值表示为 '123'::bigint；只剥离该列自身的无长度类型 cast。
		actual = strings.TrimSuffix(actual, "::"+current.typeSQL)
		return equalPostgresNumericDefaults(strings.Trim(actual, "'"), wantedSQL)
	case schema.Time, schema.Bool:
		// 未加引号的函数名/关键字不区分大小写；字符串参数和带引号的标识符必须逐字比较。
		if strings.ContainsAny(actual, "'\"") || strings.ContainsAny(wantedSQL, "'\"") {
			return false
		}
		return strings.EqualFold(actual, wantedSQL)
	default:
		return false
	}
}

var postgresNumericLiteral = regexp.MustCompile(`^[+-]?(?:[0-9]+(?:\.[0-9]*)?|\.[0-9]+)(?:[eE][+-]?[0-9]+)?$`)

func equalPostgresNumericDefaults(a, b string) bool {
	// 只比较数值字面量，不能把 SQL 的 1/2 当作有理数来计算；也不能用 float64 舍入大整数。
	if !postgresNumericLiteral.MatchString(a) || !postgresNumericLiteral.MatchString(b) {
		return false
	}
	x, okX := new(big.Rat).SetString(a)
	y, okY := new(big.Rat).SetString(b)
	return okX && okY && x.Cmp(y) == 0
}
