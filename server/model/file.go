package model

type Files struct {
	TypeId    uint   `json:"type_id" gorm:"type:tinyint UNSIGNED not null;default:0;comment:分类id"`        //分类id
	Uid       uint   `json:"uid" gorm:"type:int UNSIGNED not null;default:0;comment:用户id"`                //用户id
	From      string `json:"from" gorm:"type:varchar(255) not null;default:'';comment:用户来源"`              //用户来源
	Name      string `json:"name" gorm:"type:varchar(255) not null;default:'';comment:文件名"`               // 文件名
	Url       string `json:"url" gorm:"type:varchar(255) not null;default:'';comment:文件地址"`               // 文件地址
	Tag       string `json:"tag" gorm:"type:varchar(255) not null;default:'';comment:文件标签"`               // 文件标签
	Key       string `json:"key" gorm:"type:varchar(255) not null;default:'';comment:编号"`                 // 编号
	TempExist bool   `json:"temp_exist" gorm:"type:tinyint UNSIGNED not null;default:0;comment:临时文件是否存在"` // 临时文件是否存在
	Author    string `json:"author" gorm:"-"`
	PhotoId   uint   `json:"photo_id" gorm:"-"`
	TableBase
}
type FilesTemp struct {
	PhotoId uint   `json:"photo_id" gorm:"type:int UNSIGNED not null;default:0;comment:源照片id"` //源照片id
	Url     string `json:"url" gorm:"type:varchar(255) not null;default:'';comment:文件地址"`      //文件地址
	Key     string `json:"key" gorm:"type:varchar(255) not null;default:'';comment:编号"`        // 编号
	TableBase
}

func NewFile(where ...interface{}) *Files {
	var t = new(Files)
	if len(where) > 0 {
		t.DB = DB.Table("files").Model(&Files{}).Where(where[0], where[1:]...)
	} else {
		t.DB = DB.Table("files").Model(&Files{})
	}
	return t
}
func NewFileTemp(where ...interface{}) *FilesTemp {
	var t = new(FilesTemp)
	if len(where) > 0 {
		t.DB = DB.Table("files_temps").Model(&FilesTemp{}).Where(where[0], where[1:]...)
	} else {
		t.DB = DB.Table("files_temps").Model(&FilesTemp{})
	}
	return t
}
