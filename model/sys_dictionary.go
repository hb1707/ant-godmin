package model

type SysDictionary struct {
	Name                 string                `json:"name" form:"name" gorm:"column:name;comment:字典名（中）"`               // 字典名（中）
	Type                 string                `json:"type" form:"type" gorm:"column:type;comment:字典名（英）"`               // 字典名（英）
	Status               bool                  `json:"status" form:"status" gorm:"column:status;comment:状态;default:false"` // 状态
	Desc                 string                `json:"desc" form:"desc" gorm:"column:description;comment:描述"`              // 描述
	SysDictionaryDetails []SysDictionaryDetail `json:"sysDictionaryDetails" form:"sysDictionaryDetails"`
	TableBase
}
type SysDictionaryDetail struct {
	Label           string `json:"label" form:"label" gorm:"column:label;comment:展示值"`                                   // 展示值
	Value           int    `json:"value" form:"value" gorm:"column:value;comment:字典值;type:integer"`                      // 字典值
	Status          bool   `json:"status" form:"status" gorm:"column:status;comment:启用状态;default:false"`                // 启用状态
	Sort            int    `json:"sort" form:"sort" gorm:"column:sort;comment:排序标记;type:integer"`                       // 排序标记
	SysDictionaryID uint   `json:"sysDictionaryID" form:"sysDictionaryID" gorm:"column:sys_dictionary_id;comment:关联标记"` // 关联标记
	TableBase
}

func NewSysDictionary(where ...interface{}) *SysDictionary {
	var t = new(SysDictionary)
	if len(where) > 0 {
		t.DB = DB.Table("sys_dictionaries").Model(&SysDictionary{}).Where(where[0], where[1:]...)
	} else {
		t.DB = DB.Table("sys_dictionaries").Model(&SysDictionary{})
	}
	return t
}
func NewDictionaryDetail(where ...interface{}) *SysDictionaryDetail {
	var t = new(SysDictionaryDetail)
	if len(where) > 0 {
		t.DB = DB.Table("sys_dictionary_details").Model(&SysDictionaryDetail{}).Where(where[0], where[1:]...)
	} else {
		t.DB = DB.Table("sys_dictionary_details").Model(&SysDictionaryDetail{})
	}
	return t
}
func (t *SysDictionary) All() ([]SysDictionary, []SysDictionaryDetail) {
	var list []SysDictionary
	var listDetail []SysDictionaryDetail
	err = t.DB.Find(&list).Error
	if failed(err) {
		return []SysDictionary{}, nil
	}
	err = NewDictionaryDetail().DB.Where("status = ?", true).Find(&listDetail).Order("sort asc").Error
	if failed(err) {
		return []SysDictionary{}, nil
	}
	return list, listDetail
}
