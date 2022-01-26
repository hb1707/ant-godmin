package model

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ReqPageSize struct {
	Page int `json:"page" form:"page"`
	Size int `json:"size" form:"size"`
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
	err := c.ShouldBindJSON(&req)
	if err != nil {
		t.Limit = defaultSize
		t.Page = 0
	} else {
		if req.Size > 0 {
			t.Limit = req.Size
		} else {
			t.Limit = defaultSize
		}
		t.Page = req.Page
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
			err = t.DB.Where("id", t.Id).Updates(t.Data).Error
		} else {
			err = t.DB.Create(t.Data).Error
		}
	} else {
		if t.Id > 0 {
			err = t.DB.Where("id", t.Id).Select(must[0], must[1:]...).Updates(t.Req).Error
		} else {
			err = t.DB.Create(t.Req).Error
		}
	}
	if failed(err) {
		return err
	}
	return nil
}
