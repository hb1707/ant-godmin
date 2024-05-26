package json

import (
	"github.com/gin-gonic/gin"
	"github.com/hb1707/ant-godmin/auth"
	"github.com/hb1707/ant-godmin/cache"
	"github.com/hb1707/ant-godmin/model"
	"net/http"
	"strconv"
)

func FetchOne(c *gin.Context) {
	var data = map[string]interface{}{}

	table := c.Param("table")
	var tables = cache.Tables(table, false)
	if tables.TableName == "" {
		jsonErr(c, http.StatusNotFound, ErrorEmpty)
		return
	}
	var fields = cache.TableFields(table, false)
	var req = make(map[string]any)
	for _, field := range fields {
		if v, exist := c.GetQuery(field.InputName); exist {
			if field.Type == "int" {
				req[field.InputName], _ = strconv.Atoi(v)
			} else if field.Type == "string" {
				req[field.InputName] = v
			} else if field.Type == "bool" {
				req[field.InputName], _ = strconv.ParseBool(v)
			} else if field.Type == "float" {
				req[field.InputName], _ = strconv.ParseFloat(v, 64)
			} else {
				req[field.InputName] = v
			}
		}
	}
	// 查询数据
	tb := model.NewTable(table)
	if len(req) > 0 {
		tb.Where(req)
	}
	tb.DB.Limit(1)
	tb.List(&data)
	jsonResult(c, http.StatusOK, data)
	return
}
func FetchAll(c *gin.Context) {
	var data []map[string]interface{}
	table := c.Param("table")
	var tables = cache.Tables(table, false)
	if tables.TableName == "" {
		jsonErr(c, http.StatusNotFound, ErrorEmpty)
		return
	}
	var fields = cache.TableFields(table, false)
	var req = make(map[string]any)
	for _, field := range fields {
		if v, exist := c.GetQuery(field.InputName); exist {
			if field.Type == "int" {
				req[field.InputName], _ = strconv.Atoi(v)
			} else if field.Type == "string" {
				req[field.InputName] = v
			} else if field.Type == "bool" {
				req[field.InputName], _ = strconv.ParseBool(v)
			} else if field.Type == "float" {
				req[field.InputName], _ = strconv.ParseFloat(v, 64)
			} else {
				req[field.InputName] = v
			}
		}
	}
	// 查询数据
	tb := model.NewTable(table)
	if len(req) > 0 {
		tb.Where(req)
	}
	tb.PageAndLimit(c)
	tb.List(&data)
	jsonResult(c, http.StatusOK, data)
	return
}
func Create(c *gin.Context) {
	// 获取表名
	table := c.Param("table")
	uid, authID := auth.Identity(c)
	if authID == "" || uid == 0 {
		jsonErr(c, http.StatusUnauthorized, ErrorPermission)
		return
	}
	// 获取请求参数
	var fields = cache.TableFields(table, false)
	var req = make(map[string]any)
	for _, field := range fields {
		if v, exist := c.GetPostForm(field.InputName); exist {
			if field.Type == "int" {
				req[field.InputName], _ = strconv.Atoi(v)
			} else if field.Type == "string" {
				req[field.InputName] = v
			} else if field.Type == "bool" {
				req[field.InputName], _ = strconv.ParseBool(v)
			} else if field.Type == "float" {
				req[field.InputName], _ = strconv.ParseFloat(v, 64)
			} else {
				req[field.InputName] = v
			}
		}
	}
	// 创建数据
	tb := model.NewTable(table)
	tb.DB.Create(req)
	jsonResult(c, http.StatusOK, req)
	return
}
func Update(c *gin.Context) {
	// 获取表名
	table := c.Param("table")
	id := c.Param("id")
	if id == "" {
		jsonErr(c, http.StatusBadRequest, ErrorParameterID)
		return
	}
	uid, authID := auth.Identity(c)
	if authID == "" || uid == 0 {
		jsonErr(c, http.StatusUnauthorized, ErrorPermission)
		return
	}
	// 获取请求参数
	var fields = cache.TableFields(table, false)
	var req = make(map[string]any)
	for _, field := range fields {
		if v, exist := c.GetPostForm(field.InputName); exist {
			if field.Type == "int" {
				req[field.InputName], _ = strconv.Atoi(v)
			} else if field.Type == "string" {
				req[field.InputName] = v
			} else if field.Type == "bool" {
				req[field.InputName], _ = strconv.ParseBool(v)
			} else if field.Type == "float" {
				req[field.InputName], _ = strconv.ParseFloat(v, 64)
			} else {
				req[field.InputName] = v
			}
		}
	}
	// 更新数据
	tb := model.NewTable(table)
	tb.DB.Where("id = ?", id).Updates(req)
	jsonResult(c, http.StatusOK, req)
	return
}

func Delete(c *gin.Context) {
	// 获取表名
	table := c.Param("table")
	id := c.Param("id")
	if id == "" {
		jsonErr(c, http.StatusBadRequest, ErrorParameterID)
		return
	}
	uid, authID := auth.Identity(c)
	if authID == "" || uid == 0 {
		jsonErr(c, http.StatusUnauthorized, ErrorPermission)
		return
	}
	// 删除数据
	tb := model.NewTable(table)
	tb.DB.Where("id = ?", id).Delete(id)
	jsonResult(c, http.StatusOK, nil)
	return
}
