package model

type Files struct {
	TypeId    uint   `json:"type_id" gorm:"comment:分类id"`  //分类id
	Uid       uint   `json:"uid" gorm:"comment:用户id"`      //用户id
	From      string `json:"from" gorm:"comment:用户来源"`     //用户来源
	Name      string `json:"name" gorm:"comment:文件名"`      // 文件名
	Url       string `json:"url" gorm:"comment:文件地址"`      // 文件地址
	Tag       string `json:"tag" gorm:"comment:文件标签"`      // 文件标签
	Key       string `json:"key" gorm:"comment:编号"`        // 编号
	TempExist bool   `json:"temp_exist" gorm:"comment:编号"` // 编号
	Author    string `json:"author" gorm:"-"`
	PhotoId   uint   `json:"photo_id" gorm:"-"`
	TableBase
}
type FilesTemp struct {
	PhotoId uint   `json:"photo_id" gorm:"comment:源照片id"` //源照片id
	Url     string `json:"url" gorm:"comment:文件地址"`       // 文件地址
	Key     string `json:"key" gorm:"comment:编号"`         // 编号
	TableBase
}

func NewFile(where ...interface{}) *Files {
	var t = new(Files)
	if len(where) > 0 {
		t.DB = DB.Table("wxc_photos").Model(&Files{}).Where(where[0], where[1:]...)
	} else {
		t.DB = DB.Table("wxc_photos").Model(&Files{})
	}
	return t
}
func NewFileTemp(where ...interface{}) *FilesTemp {
	var t = new(FilesTemp)
	if len(where) > 0 {
		t.DB = DB.Table("wxc_photos_temps").Model(&FilesTemp{}).Where(where[0], where[1:]...)
	} else {
		t.DB = DB.Table("wxc_photos_temps").Model(&FilesTemp{})
	}
	return t
}
