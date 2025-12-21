package json

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hb1707/ant-godmin/cache"
	"github.com/hb1707/ant-godmin/model"
	"github.com/hb1707/ant-godmin/setting"
	"net/http"
	"strconv"
)

func FetchTablesAll(c *gin.Context) {
	var tables []model.Tables
	tables = model.NewTables().All("sort asc")
	jsonResult(c, http.StatusOK, tables)
	return
}

func DetailTable(c *gin.Context) {
	idStr := c.Query("id")
	id, _ := strconv.Atoi(idStr)
	var exist = model.NewTables("id = ?", id).GetOne("sort asc")
	if exist.TableName == "" {
		jsonErr(c, http.StatusBadRequest, ErrorEmpty)
		return
	}
	data := model.TablesSqlToForm(*exist)
	jsonResult(c, http.StatusOK, data)
	return
}

func EditTables(c *gin.Context) {
	var req model.TablesForm
	err := c.ShouldBindJSON(&req)
	if err != nil {
		jsonErr(c, http.StatusBadRequest, ErrorParameter)
		return
	}
	var exist = model.NewTables("id = ? OR table_name = ?", req.Id, req.InputName).GetOne("sort asc")
	table := model.NewTables()
	if exist.TableName != "" {
		table.Id = exist.Id
		//修改表名
		if setting.DB.DRIVER == "postgres" {
			model.DB.Exec(fmt.Sprintf("ALTER TABLE %s RENAME TO %s", setting.DB.PRE+exist.TableName, req.InputName))
		} else {
			model.DB.Exec(fmt.Sprintf("ALTER TABLE `%s` RENAME TO `%s`", setting.DB.PRE+exist.TableName, req.InputName))
		}
	} else {
		//新增表
		if setting.DB.DRIVER == "postgres" {
			model.DB.Exec(fmt.Sprintf("CREATE TABLE %s (id SERIAL PRIMARY KEY, created_at timestamp with time zone, updated_at timestamp with time zone, deleted_at timestamp with time zone)", setting.DB.PRE+req.InputName))
		} else {
			model.DB.Exec(fmt.Sprintf("CREATE TABLE `%s` (`id` int NOT NULL AUTO_INCREMENT, `created_at` datetime NULL DEFAULT NULL, `updated_at` datetime NULL DEFAULT NULL, `deleted_at` datetime NULL DEFAULT NULL, PRIMARY KEY (`id`) USING BTREE) ", setting.DB.PRE+req.InputName))
		}
	}
	table.TableName = req.InputName
	table.Label = req.InputLabel
	if req.UploadImage != nil {
		table.Image = *req.UploadImage
	}
	if req.InputDesc != nil {
		table.Desc = *req.InputDesc
	}
	table.Role = req.Role
	table.Sort = req.Sort
	table.Edit()
	cache.Tables(table.TableName, true)
	jsonResult(c, http.StatusOK, table)
	return
}

func DelTables(c *gin.Context) {
	idStr := c.Query("id")
	id, _ := strconv.Atoi(idStr)
	var exist = model.NewTables("id = ?", id).GetOne("sort asc")
	if exist.TableName == "" {
		jsonErr(c, http.StatusNotFound, ErrorEmpty)
		return
	}
	//model.DB.Exec(fmt.Sprintf("DROP TABLE `%s`", exist.TableName))
	model.NewTables("id = ?", id).Del(exist.Id)
	cache.Tables(exist.TableName, true)
	jsonResult(c, http.StatusOK, nil)
	return
}
