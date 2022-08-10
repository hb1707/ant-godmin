module github.com/hb1707/ant-godmin

go 1.18

require (
	github.com/aliyun/aliyun-oss-go-sdk v2.2.2+incompatible
	github.com/appleboy/gin-jwt/v2 v2.7.0
	github.com/flosch/pongo2/v4 v4.0.2
	github.com/gin-contrib/cors v1.3.1
	github.com/gin-gonic/gin v1.7.4
	github.com/go-ini/ini v1.63.2
	github.com/hb1707/exfun v0.0.0-20210929093952-21a6a403b4c4
	github.com/satori/go.uuid v1.2.0
	github.com/silenceper/wechat/v2 v2.0.9
	gorm.io/driver/mysql v1.1.2
	gorm.io/gorm v1.21.16
)

require (
	github.com/baiyubin/aliyun-sts-go-sdk v0.0.0-20180326062324-cfa1a18b161f // indirect
	github.com/bradfitz/gomemcache v0.0.0-20190913173617-a41fca850d0b // indirect
	github.com/fatih/structs v1.1.0 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-playground/locales v0.13.0 // indirect
	github.com/go-playground/universal-translator v0.17.0 // indirect
	github.com/go-playground/validator/v10 v10.4.1 // indirect
	github.com/go-sql-driver/mysql v1.6.0 // indirect
	github.com/golang-jwt/jwt/v4 v4.1.0 // indirect
	github.com/golang/protobuf v1.3.3 // indirect
	github.com/gomodule/redigo v1.8.4 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.2 // indirect
	github.com/json-iterator/go v1.1.9 // indirect
	github.com/leodido/go-urn v1.2.0 // indirect
	github.com/mattn/go-isatty v0.0.12 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/ugorji/go/codec v1.1.7 // indirect
	golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83 // indirect
	golang.org/x/sys v0.0.0-20210309074719-68d13333faf2 // indirect
	golang.org/x/time v0.0.0-20220411224347-583f2d630306 // indirect
	gopkg.in/yaml.v2 v2.2.8 // indirect
)

replace github.com/silenceper/wechat/v2 => ../../../../DEV/wechat

replace github.com/hb1707/exfun => ../../../../DEV/exfun
