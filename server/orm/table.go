package orm

import "gorm.io/gorm"

func (t *TableBase) Request(data interface{}) *TableBase {
    t.Req = data
    return t
}
func (t *TableBase) PageAndLimit(page, limit int) *TableBase {
    t.Limit = limit
    t.Page = page
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

func (t *TableBase) AddOrUpdate() {
    if t.Data != nil {
        if t.Id > 0 {
            err = t.DB.Where("id", t.Id).Updates(t.Data).Error
        } else {
            err = t.DB.Create(t.Data).Error
        }
    } else {
        if t.Id > 0 {
            err = t.DB.Where("id", t.Id).Updates(t.Req).Error
        } else {
            err = t.DB.Create(t.Req).Error
        }
    }
    if failed(err) {
        return
    }
    return
}
