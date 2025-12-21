package model

import "github.com/hb1707/ant-godmin/pkg/log"

type FieldType string

const (
	FieldTypeString FieldType = "string"
	FieldTypeInt    FieldType = "int"
	FieldTypeFloat  FieldType = "float"
	FieldTypeBool   FieldType = "bool"
	FieldTypeTime   FieldType = "datetime"
	FieldTypeText   FieldType = "text"
	FieldTypeJson   FieldType = "json"
	FieldTypeFile   FieldType = "file"
	FieldTypeImage  FieldType = "image"
)

var FieldTypeMap = map[FieldType]string{
	FieldTypeString: "varchar",
	FieldTypeInt:    "int",
	FieldTypeFloat:  "float",
	FieldTypeBool:   "boolean",
	FieldTypeTime:   "timestamp",
	FieldTypeText:   "text",
	FieldTypeJson:   "jsonp",
	FieldTypeFile:   "varchar",
	FieldTypeImage:  "varchar",
}

func init() {
	if confDB.DRIVER == "mysql" {
		FieldTypeMap[FieldTypeBool] = "tinyint"
		FieldTypeMap[FieldTypeTime] = "datetime"
		FieldTypeMap[FieldTypeJson] = "json"
	}
}

type Fields struct {
	TableName    string    `json:"tableName" gorm:"column:table_name;type:varchar(100);not null;default:'';"`       // table
	Label        string    `json:"label" gorm:"column:label;type:varchar(100);not null;default:'';"`                // label
	FieldName    string    `json:"fieldName" gorm:"column:field_name;type:varchar(100);not null;default:'';"`       // name
	FieldType    FieldType `json:"fieldType" gorm:"column:field_type;type:varchar(100);not null;default:'';"`       // type
	Role         string    `json:"role" gorm:"column:role;type:varchar(100);not null;default:'100';"`               // role
	Sort         int       `json:"sort" gorm:"column:sort;not null;default:0;type:integer;"`                        // sort
	MinRequired  int       `json:"minRequired" gorm:"column:min_required;not null;default:0;type:integer;"`         // min
	MaxRequired  int       `json:"maxRequired" gorm:"column:max_required;not null;default:0;type:integer;"`         // max
	AllowSearch  bool      `json:"allowSearch" gorm:"column:allow_search;not null;default:false;"`                  // allow search
	DefaultValue string    `json:"defaultValue" gorm:"column:default_value;type:varchar(100);not null;default:'';"` // default value
	Tips         string    `json:"tips" gorm:"column:tips;type:varchar(100);not null;default:'';"`                  // tips
	TextRegexp   string    `json:"textRegexp" gorm:"column:text_regexp;type:varchar(255);not null;default:'';"`     // text regexp
	IsUnique     bool      `json:"isUnique" gorm:"column:is_unique;not null;default:false;"`                        // is unique
	IsPrivate    bool      `json:"isPrivate" gorm:"column:is_private;not null;default:false;"`                      // is private
	IsRequired   bool      `json:"isRequired" gorm:"column:is_required;not null;default:false;"`                    // is required
	TableBase
}

type FormField struct {
	Id              uint      `json:"id"`
	Type            FieldType `json:"type" required:"true"`
	Role            string    `json:"role"`
	InputLabel      string    `json:"input_label"`
	InputName       string    `json:"input_name"`
	InputTips       string    `json:"input_tips"`
	InputRegexp     string    `json:"input_regexp"`
	InputMin        *int      `json:"input_min"`
	InputMax        *int      `json:"input_max"`
	InputDefault    string    `json:"input_default"`
	AllowSearchable bool      `json:"allow_searchable"`
	IsPrivate       bool      `json:"is_private"`
	IsRequired      bool      `json:"is_required"`
	IsUnique        bool      `json:"is_unique"`
}

func FieldSqlToForm(f Fields) FormField {

	var tableField = FormField{
		Id:              f.Id,
		Type:            f.FieldType,
		Role:            f.Role,
		InputLabel:      f.Label,
		InputName:       f.FieldName,
		InputTips:       f.Tips,
		InputRegexp:     f.TextRegexp,
		InputDefault:    f.DefaultValue,
		AllowSearchable: f.AllowSearch,
		IsPrivate:       f.IsPrivate,
		IsRequired:      f.IsRequired,
		IsUnique:        f.IsUnique,
	}
	if f.MinRequired != 0 {
		tableField.InputMin = &f.MinRequired
	}
	if f.MaxRequired != 0 {
		tableField.InputMax = &f.MaxRequired
	}
	return tableField
}

func NewFields(where ...interface{}) *Fields {
	var t = new(Fields)
	if len(where) > 0 {
		t.DB = DB.Model(&Fields{}).Where(where[0], where[1:]...)
	} else {
		t.DB = DB.Model(&Fields{})
	}
	return t
}
func (t *Fields) All(order string) []Fields {
	var list []Fields
	t.List(&list, order)
	return list
}

func (t *Fields) GetOne(order string) *Fields {
	var user Fields
	t.One(&user, order)
	return &user
}

func (t *Fields) Edit(must []string) *Fields {
	var user Fields
	t.Request(t)
	err := t.AddOrUpdate(must)
	if err != nil {
		log.Error(err)
	}
	return &user
}
