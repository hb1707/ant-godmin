package model

import "github.com/hb1707/ant-godmin/pkg/log"

type Tables struct {
	TableName string `json:"tableName" gorm:"column:table_name;type:varchar(100);not null;"` // table
	Label     string `json:"label" gorm:"column:label;type:varchar(100);not null;"`          // label
	Image     string `json:"image" gorm:"column:image;type:varchar(255);not null;"`          // image
	Desc      string `json:"desc" gorm:"column:desc;type:text;"`                             // desc
	Role      string `json:"role" gorm:"column:role;type:varchar(100);not null;"`            // role
	Sort      int    `json:"sort" gorm:"column:sort;not null;type:integer;"`                 // sort
	TableBase
}
type TablesForm struct {
	Id          uint    `json:"id"`
	Role        string  `json:"role"`
	Sort        int     `json:"sort"`
	InputName   string  `json:"input_name"`
	InputLabel  string  `json:"input_label"`
	UploadImage *string `json:"upload_image"`
	InputDesc   *string `json:"input_desc"`
}

func TablesSqlToForm(t Tables) TablesForm {
	return TablesForm{
		Id:          t.Id,
		Role:        t.Role,
		Sort:        t.Sort,
		InputName:   t.TableName,
		InputLabel:  t.Label,
		UploadImage: &t.Image,
		InputDesc:   &t.Desc,
	}
}

func NewTables(where ...interface{}) *Tables {
	var t = new(Tables)
	if len(where) > 0 {
		t.DB = DB.Model(&Tables{}).Where(where[0], where[1:]...)
	} else {
		t.DB = DB.Model(&Tables{})
	}
	return t
}
func (t *Tables) All(order string) []Tables {
	var list []Tables
	t.List(&list, order)
	return list
}

func (t *Tables) GetOne(order string) *Tables {
	var user Tables
	t.One(&user, order)
	return &user
}

func (t *Tables) Edit() *Tables {
	var user Tables
	t.Request(t)
	err := t.AddOrUpdate()
	if err != nil {
		log.Error(err)
	}
	return &user
}
