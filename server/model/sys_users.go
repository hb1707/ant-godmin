package model

import (
	"github.com/satori/go.uuid"
)

type SysUsers struct {
	UUID        uuid.UUID `json:"uuid" gorm:"comment:用户UUID"`                                                    // 用户UUID
	Username    string    `json:"userName" gorm:"comment:用户登录名"`                                                 // 用户登录名
	Password    string    `json:"-"  gorm:"comment:用户登录密码"`                                                      // 用户登录密码
	RealName    string    `json:"realName" gorm:"default:系统用户;comment:用户实名"`                                     // 用户昵称
	NickName    string    `json:"nickName" gorm:"default:系统用户;comment:用户昵称"`                                     // 用户昵称
	HeaderImg   string    `json:"avatar" gorm:"default:https://adm.wanxuechuang.com/data/head.png;comment:用户头像"` // 用户头像
	AuthorityId string    `json:"authorityId" gorm:"default:999;comment:用户角色ID"`                                 // 用户角色ID
	SideMode    string    `json:"sideMode" gorm:"default:dark;comment:用户角色ID"`                                   // 用户侧边主题
	ActiveColor string    `json:"activeColor" gorm:"default:#1890ff;comment:用户角色ID"`                             // 活跃颜色
	BaseColor   string    `json:"baseColor" gorm:"default:#fff;comment:用户角色ID"`                                  // 基础颜色
	TypeId      int       `json:"typeId" form:"typeId" gorm:"column:post_type;type:smallint"`                    //岗位类型
	Salt        string    `json:"-" gorm:"type:varchar(10) not null"`                                            //salt
	QywxUserid  string    `json:"-" gorm:"type:varchar(100) not null"`                                           //企业微信
	WxUnionId   string    `json:"-" gorm:"type:varchar(100) not null"`                                           //微信
	XmUserId    int       `json:"-" form:"xm_user_id" gorm:"column:xm_user_id;type:bigint"`
	TableBase
}

func NewSysUser(where ...interface{}) *SysUsers {
	var t = new(SysUsers)
	if len(where) > 0 {
		t.DB = DB.Table("sys_users").Model(&SysUsers{}).Where(where[0], where[1:]...)
	} else {
		t.DB = DB.Table("sys_users").Model(&SysUsers{})
	}
	return t
}
func (t *SysUsers) All(order string) []SysUsers {
	var list []SysUsers
	if order != "" {
		err = t.DB.Order(order).Find(&list).Error
	} else {
		err = t.DB.Find(&list).Error
	}
	if failed(err) {
		return []SysUsers{}
	}
	return list
}

func (t *SysUsers) One(order string) *SysUsers {
	if order != "" {
		err = t.DB.Order(order).First(t).Error
	} else {
		err = t.DB.First(t).Error
	}
	if failed(err) {
		return new(SysUsers)
	}
	return t
}
func (t *SysUsers) DataMap(data map[string]interface{}) {
	t.Data = data
}

func (t *SysUsers) AddOrUpdate() *SysUsers {
	if t.Data != nil {
		if t.Id > 0 {
			err = t.DB.Where("id", t.Id).Updates(t.Data).Error
		} else {
			err = t.DB.Create(t.Data).Error
		}
	} else {
		if t.Id > 0 {
			err = t.DB.Where("id", t.Id).Updates(t).Error
		} else {
			err = t.DB.Create(t).Error
		}
	}
	if failed(err) {
		return new(SysUsers)
	}
	return t
}
