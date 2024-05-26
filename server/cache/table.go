package cache

import "github.com/hb1707/ant-godmin/model"

var tableFieldsCache = make(map[string][]model.TableField)

func TableFields(table string, forceUpdate bool) []model.TableField {
	var fields []model.Fields
	if _fields, exist := tableFieldsCache[table]; exist && len(_fields) > 0 && !forceUpdate {
		return _fields
	} else {
		fields = model.NewFields("table_name = ? ", table).All("sort asc")
		tableFieldsCache[table] = make([]model.TableField, len(fields))
		for i, field := range fields {
			tableFieldsCache[table][i] = model.FieldSqlToJson(field)
		}
		return tableFieldsCache[table]
	}

}

var tablesCache = make(map[string]model.Tables)

func Tables(table string, forceUpdate bool) model.Tables {
	var tables []model.Tables
	if _tables, exist := tablesCache[table]; exist && _tables.TableName != "" && !forceUpdate {
		return _tables
	} else {
		tables = model.NewTables().All("sort asc")
		for _, table := range tables {
			tablesCache[table.TableName] = table
		}
		return tablesCache[table]
	}
}
