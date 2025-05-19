package model

import (
	"github.com/hb1707/ant-godmin/consts"
	"gorm.io/gorm"
	"strings"
)

type Files struct {
	TypeId    uint             `json:"type_id" gorm:"type:tinyint UNSIGNED not null;default:0;comment:分类id"`              //分类id 0 图片 2其他 3证件 4APK
	CloudType consts.CloudType `json:"cloud_type" gorm:"type:tinyint UNSIGNED not null;default:0;comment:云类型"`           //云类型
	FileType  consts.FileType  `json:"file_type" gorm:"type:tinyint UNSIGNED not null;default:0;comment:文件类型"`          //文件类型,0 未知，1 图片，2 JSON
	Domain    string           `json:"domain" gorm:"type:varchar(64) not null;default:'';comment:域名"`                     //域名
	UserSpace string           `json:"user_space" gorm:"type:varchar(64) not null;default:'';comment:用户空间"`             //用户空间
	Uid       uint             `json:"uid" gorm:"type:int UNSIGNED not null;default:0;comment:用户id"`                      //用户id
	From      string           `json:"from" gorm:"type:varchar(255) not null;default:'';comment:用户来源"`                  //用户来源
	Name      string           `json:"name" gorm:"type:varchar(255) not null;default:'';comment:文件名"`                    // 文件名
	Url       string           `json:"url" gorm:"type:varchar(255) not null;default:'';comment:文件地址"`                   // 文件地址
	Tag       string           `json:"tag" gorm:"type:varchar(255) not null;default:'';comment:文件标签"`                   // 文件标签
	Key       string           `json:"key" gorm:"type:varchar(255) not null;default:'';comment:编号"`                       // 编号
	TempExist bool             `json:"temp_exist" gorm:"type:tinyint UNSIGNED not null;default:0;comment:临时文件是否存在"` // 临时文件是否存在
	Other     FileOther        `json:"other" gorm:"type:json;comment:其他信息"`                                             //其他信息
	Content   string           `json:"content" gorm:"type:longtext;comment:文件内容"`                                       //文件内容
	Author    string           `json:"author" gorm:"-"`
	FileId    uint             `json:"file_id" gorm:"-"`
	TableBase
}
type FilesTemp struct {
	FileId uint   `json:"file_id" gorm:"type:int UNSIGNED not null;default:0;comment:源文件id"` //源文件id
	Url    string `json:"url" gorm:"type:varchar(255) not null;default:'';comment:文件地址"`    //文件地址
	Key    string `json:"key" gorm:"type:varchar(255) not null;default:'';comment:编号"`        // 编号
	TableBase
}

func NewFile(where ...interface{}) *Files {
	var t = new(Files)
	if len(where) > 0 {
		t.DB = DB.Model(&Files{}).Where(where[0], where[1:]...)
	} else {
		t.DB = DB.Model(&Files{})
	}
	return t
}

func (t *Files) AfterFind(tx *gorm.DB) (err error) {
	if t.Domain != "" {
		t.Url = strings.Replace(t.Url, "{DOMAIN}", t.Domain, 1)
	}
	return
}

func NewFileTemp(where ...interface{}) *FilesTemp {
	var t = new(FilesTemp)
	if len(where) > 0 {
		t.DB = DB.Model(&FilesTemp{}).Where(where[0], where[1:]...)
	} else {
		t.DB = DB.Model(&FilesTemp{})
	}
	return t
}
