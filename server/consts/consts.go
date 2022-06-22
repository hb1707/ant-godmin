package consts

import "errors"

const AccessToken = "Access-Token"
const ApplicationJson = "application/json"
const ApplicationUrlencoded = "application/x-www-form-urlencoded; charset=UTF-8"
const ApplicationXml = "application/xml;charset=utf-8"

var (
	ErrAlreadyRegistered    = errors.New("帐号已经被占用，请勿重复注册！")
	ErrMissingRegister      = errors.New("注册时缺少参数或者参数类型错误！")
	ErrEmptyMoblie          = errors.New("请先确保您的企业微信资料里面已经填写了手机号！")
	ErrInconsistentPassword = errors.New("请确保两个密码一致！")
	ErrMissingLoginValues   = errors.New("缺少帐号或密码")
	ErrFailedAuthentication = errors.New("帐号或密码错误")
	ErrFailedCode           = errors.New("帐号或验证码错误")
	ErrEmptyUnionID         = errors.New("UnionID为空，未绑定微信开放平台")
	ErrMissingParameter     = errors.New("缺少参数")
	ErrUnauthorized         = errors.New("非法访问")
	ErrMissingDomain        = errors.New("域名没有配置，请检查/configs/.env")
	ErrNotFound             = errors.New(`内容已被删除~`)
	ErrJson                 = errors.New(`JSON解析错误！`)
	ErrUnregistered         = errors.New(`unregistered`)
)

const AuthorityIdAdmin string = "101"

type FileType int

const (
	FileTypeOther FileType = 0
	FileTypeImage FileType = 1
	FileTypeJson  FileType = 2
)
