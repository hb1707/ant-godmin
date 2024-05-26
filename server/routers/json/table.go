package json

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hb1707/ant-godmin/cache"
	"github.com/hb1707/ant-godmin/model"
	"github.com/hb1707/ant-godmin/setting"
	"github.com/hb1707/exfun/fun"
	"net/http"
	"strconv"
)

func FetchTablesAll(c *gin.Context) {
	var tables []model.Tables
	tables = model.NewTables().All("sort asc")
	jsonResult(c, http.StatusOK, tables)
	return

}

type reqTable struct {
	TableName string `json:"tableName"`
	Label     string `json:"label"`
	Role      string `json:"role"`
	Sort      int    `json:"sort"`
}

func EditTables(c *gin.Context) {
	var req reqTable
	err := c.ShouldBindJSON(&req)
	if err != nil {
		jsonErr(c, http.StatusBadRequest, ErrorParameter)
		return
	}
	var exist = model.NewTables("table_name = ?", req.TableName).GetOne("sort asc")
	table := model.NewTables()
	if exist.TableName != "" {
		table.Id = exist.Id
		//修改表名
		model.DB.Exec(fmt.Sprintf("ALTER TABLE `%s` RENAME TO `%s`", setting.DB.PRE+exist.TableName, req.TableName))
	} else {
		//新增表
		model.DB.Exec(fmt.Sprintf("CREATE TABLE `%s` (`id` int NOT NULL AUTO_INCREMENT, `created_at` datetime NULL DEFAULT NULL, `updated_at` datetime NULL DEFAULT NULL, `deleted_at` datetime NULL DEFAULT NULL, PRIMARY KEY (`id`) USING BTREE) ", setting.DB.PRE+req.TableName))
	}
	table.TableName = req.TableName
	table.Label = req.Label
	table.Role = req.Role
	table.Sort = req.Sort
	table.Edit()
	cache.Tables(table.TableName, true)
	jsonResult(c, http.StatusOK, table)
	return
}

func DelTables(c *gin.Context) {
	table := c.Param("table")
	var exist = model.NewTables("table_name = ?", table).GetOne("sort asc")
	if exist.TableName == "" {
		jsonErr(c, http.StatusNotFound, ErrorEmpty)
		return
	}
	//model.DB.Exec(fmt.Sprintf("DROP TABLE `%s`", exist.TableName))
	model.NewTables("table_name = ?", table).Del(exist.Id)
	cache.Tables(table, true)
	jsonResult(c, http.StatusOK, nil)
	return
}

func ListFields(c *gin.Context) {
	table := c.Param("table")
	fields := cache.TableFields(table, false)
	jsonResult(c, http.StatusOK, fields)
	return
}
func DetailField(c *gin.Context) {
	idStr := c.Query("id")
	id, _ := strconv.Atoi(idStr)
	field := model.NewFields("id = ?", id).GetOne("sort asc")
	data := model.FieldSqlToJson(*field)
	jsonResult(c, http.StatusOK, data)
	return
}

func EditFields(c *gin.Context) {
	table := c.Param("table")
	var req model.TableField
	err := c.ShouldBindJSON(&req)
	if err != nil {
		jsonErr(c, http.StatusBadRequest, ErrorParameter)
		return
	}
	if !fun.IsEnglish(req.InputName+table, false) {
		jsonErr(c, http.StatusBadRequest, errors.New("value must be in English"))
		return
	}
	if !fun.IsChineseAndEnglish(req.InputLabel, false) {
		jsonErr(c, http.StatusBadRequest, errors.New("label must be in Chinese or English"))
		return
	}
	fieldType, ok := model.FieldTypeMap[req.Type]
	if !ok {
		jsonErr(c, http.StatusBadRequest, errors.New("type error"))
		return
	}
	var exist = model.NewFields("id = ?", req.Id).GetOne("sort asc")
	fields := model.NewFields()
	var defaultValue = ""
	var notNull = ""
	var maxLength = 64
	if req.InputMax != nil && *req.InputMax > 0 {
		maxLength = *req.InputMax
	}

	if req.Type == model.FieldTypeBool {
		defaultValue = fun.If2String(req.InputDefault == "true", "1", "0")
		notNull = "NOT NULL"
	} else if req.Type == model.FieldTypeInt {
		inputDefault, _ := strconv.Atoi(req.InputDefault)
		defaultValue = fun.If2String(req.InputDefault == "", "0", strconv.Itoa(inputDefault))
		notNull = "NOT NULL"
	} else if req.Type == model.FieldTypeFloat {
		inputDefault, _ := strconv.ParseFloat(req.InputDefault, 64)
		defaultValue = fun.If2String(req.InputDefault == "", "0", fmt.Sprintf("%f", inputDefault))
		notNull = "NOT NULL"
	} else if req.Type == model.FieldTypeString {
		defaultValue = fmt.Sprintf("'%s'", req.InputDefault)
		fieldType = fmt.Sprintf("%s(%d)", fieldType, maxLength*2)
		notNull = "NOT NULL"
	} else if req.Type == model.FieldTypeImage || req.Type == model.FieldTypeFile {
		notNull = "NOT NULL"
	}
	if fieldType == "varchar" {
		fieldType = fmt.Sprintf("%s(%d)", fieldType, 255)
	}
	fieldType = fmt.Sprintf("%s %s", fieldType, notNull)
	if defaultValue != "" {
		defaultValue = "DEFAULT " + defaultValue
	}
	if exist.Id > 0 && exist.FieldName != "" {
		fields.Id = exist.Id
		//修改字段
		model.DB.Exec(fmt.Sprintf("ALTER TABLE `%s` CHANGE COLUMN `%s` `%s` %s %s COMMENT \"%s\"",
			setting.DB.PRE+table,
			exist.FieldName,
			req.InputName,
			fieldType,
			defaultValue,
			req.InputLabel,
		))
	} else {
		//新增字段
		model.DB.Exec(fmt.Sprintf("ALTER TABLE `%s` ADD COLUMN `%s` %s %s COMMENT \"%s\"",
			setting.DB.PRE+table,
			req.InputName,
			fieldType,
			defaultValue,
			req.InputLabel,
		))
	}
	fields.TableName = table
	fields.Label = req.InputLabel
	fields.FieldName = req.InputName
	fields.FieldType = req.Type
	fields.Role = req.Role
	fields.AllowSearch = req.AllowSearchable
	fields.DefaultValue = req.InputDefault
	fields.Tips = req.InputTips
	fields.TextRegexp = req.InputRegexp
	if req.InputMin != nil {
		fields.MinRequired = *req.InputMin
	}
	if req.InputMax != nil {
		fields.MaxRequired = *req.InputMax
	}
	fields.IsUnique = req.IsUnique
	fields.IsPrivate = req.IsPrivate
	fields.IsRequired = req.IsRequired
	must := []string{
		"table_name",
		"field_name",
		"field_type",
		"label",
		"allow_search",
		"default_value",
		"tips",
		"text_regexp",
		"min_required",
		"max_required",
		"is_unique",
		"is_private",
		"is_required",
	}
	fields.Edit(must)
	cache.TableFields(table, true)
	jsonResult(c, http.StatusOK, fields)
	return
}

func DelFields(c *gin.Context) {
	table := c.Param("table")
	idStr := c.Query("id")
	id, _ := strconv.Atoi(idStr)
	var exist = model.NewFields(" id = ?", id).GetOne("sort asc")
	if exist.FieldName == "" {
		jsonErr(c, http.StatusNotFound, ErrorEmpty)
		return
	}
	//model.DB.Exec(fmt.Sprintf("ALTER TABLE `%s` DROP COLUMN `%s`", table, exist.FieldName))
	model.NewFields("id = ?", id).Del(exist.Id)
	cache.TableFields(table, true)
	jsonResult(c, http.StatusOK, nil)
	return
}
