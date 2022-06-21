package model

import (
	"github.com/hb1707/ant-godmin/pkg/log"
	"github.com/satori/go.uuid"
)

type SysUsers struct {
	UUID        uuid.UUID `json:"uuid" gorm:"type:varchar(100) not null;comment:用户UUID;"`                                                   // 用户UUID
	Username    string    `json:"userName" gorm:"type:varchar(100) not null;comment:用户登录名"`                                                 // 用户登录名
	Password    string    `json:"-"  gorm:"type:varchar(100) not null;comment:用户登录密码"`                                                      // 用户登录密码
	RealName    string    `json:"realName" gorm:"type:varchar(100) not null;default:系统用户;comment:用户实名"`                                     // 用户昵称
	NickName    string    `json:"nickName" gorm:"type:varchar(100) not null;default:系统用户;comment:用户昵称"`                                     // 用户昵称
	HeaderImg   string    `json:"avatar" gorm:"type:varchar(255) not null;default:https://adm.wanxuechuang.com/data/head.png;comment:用户头像"` // 用户头像
	AuthorityId string    `json:"authorityId" gorm:"type:varchar(10) not null;default:999;comment:用户角色ID"`                                  // 用户角色ID
	SideMode    string    `json:"sideMode" gorm:"type:varchar(10) not null;default:dark;comment:用户角色ID"`                                    // 用户侧边主题
	ActiveColor string    `json:"activeColor" gorm:"type:varchar(10) not null;default:#1890ff;comment:用户角色ID"`                              // 活跃颜色
	BaseColor   string    `json:"baseColor" gorm:"type:varchar(10) not null;default:#fff;comment:用户角色ID"`                                   // 基础颜色
	Salt        string    `json:"-" gorm:"type:varchar(10) not null"`                                                                       //salt
	QywxUserid  string    `json:"-" gorm:"type:varchar(100) not null"`                                                                      //企业微信
	WxUnionId   string    `json:"-" gorm:"type:varchar(100) not null"`                                                                      //微信
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
	t.List(&list, order)
	return list
}

func (t *SysUsers) GetOne(order string) *SysUsers {
	var user SysUsers
	t.One(&user, order)
	return &user
}

func (t *SysUsers) Edit() *SysUsers {
	var user SysUsers
	t.Request(t)
	err := t.AddOrUpdate()
	if err != nil {
		log.Error(err)
	}
	return &user
}
