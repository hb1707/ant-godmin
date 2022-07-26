package model

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ReqPageSize struct {
	Current  int `json:"current" form:"current"`
	PageSize int `json:"pageSize" form:"pageSize"`
}

func (t *TableBase) Where(where ...interface{}) *TableBase {
	if len(where) > 0 {
		t.DB = t.DB.Where(where[0], where[1:]...)
	}
	return t
}
func (t *TableBase) Request(data interface{}) *TableBase {
	t.Req = data
	return t
}
func (t *TableBase) PageAndLimit(c *gin.Context) *TableBase {
	var req ReqPageSize
	var defaultSize = 100
	var err error
	pageSize, existSize := c.Get("pageSize")
	current, existPage := c.Get("current")
	if c.Request.Method == "GET" {
		err = c.ShouldBindQuery(&req)
	} else if existSize && existPage {
		req.PageSize = pageSize.(int)
		req.Current = current.(int)
	} else {
		err = c.ShouldBindJSON(&req)
	}
	if err != nil {
		t.Limit = defaultSize
		t.Page = 0
	} else {
		if req.PageSize > 0 {
			t.Limit = req.PageSize
		} else {
			t.Limit = defaultSize
		}
		t.Page = req.Current
	}
	return t
}

func (t *TableBase) List(model interface{}, order ...string) {
	var dt *gorm.DB
	if len(order) > 0 {
		dt = t.DB.Order(order[0])
	} else {
		dt = t.DB
	}
	if t.Limit > 0 {
		if t.Page > 0 {
			t.Page--
		}
		dt = dt.Offset(t.Page * t.Limit).Limit(t.Limit)
	}
	err = dt.Find(model).Error
	if failed(err) {
		return
	}
	return
}
func (t *TableBase) Total() (total int64) {

	err = t.DB.Count(&total).Error
	if failed(err) {
		return
	}
	return
}
func (t *TableBase) One(model interface{}, order ...string) {
	if len(order) > 0 {
		err = t.DB.Order(order[0]).First(model).Error
	} else {
		err = t.DB.First(model).Error
	}
	if failed(err) {
		return
	}
	return
}

func (t *TableBase) DataMap(data map[string]interface{}) {
	t.Data = data
}

func (t *TableBase) AddOrUpdate(must ...interface{}) error {
	if t.Data != nil {
		if t.Id > 0 {
			err = t.DB.Where("id = ?", t.Id).Updates(t.Data).Error
		} else {
			err = t.DB.Create(t.Data).Error
		}
	} else {
		if t.Id > 0 {
			t.DB.Where("id = ?", t.Id)
			if len(must) > 0 && must[0] != "" {
				t.DB.Select(must[0], must[1:]...)
			}
			err = t.DB.Updates(t.Req).Error
		} else {
			if len(must) > 0 && must[0] != "" {
				t.DB.Select(must[0], must[1:]...)
			}
			err = t.DB.Create(t.Req).Error
		}
	}
	if failed(err) {
		return err
	}
	return nil
}
func (t *TableBase) Del(model interface{}) error {
	if t.Id > 0 {
		err = t.DB.Where("id = ?", t.Id).Delete(model).Error
		if failed(err) {
			return err
		}
	}
	return nil
}
func (t *TableBase) UpdateRows() int {
	if t.DB != nil {
		return int(t.DB.RowsAffected)
	}
	return 0
}
func (t *TableBase) UpdateField(id uint, field string, value interface{}) {
	if t.DB != nil {
		t.DB.Where("id = ?", id).Update(field, value)
	}
}
func (t *TableBase) UpdateExpr(id uint, field string, expr string, value interface{}) {
	if t.DB != nil {
		t.DB.Where("id = ?", id).UpdateColumn(field, gorm.Expr(expr, value))
	}
}
