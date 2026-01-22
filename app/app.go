package app

import (
	"github.com/hb1707/ant-godmin/common"
	"github.com/hb1707/ant-godmin/setting"
)

func Init(path string) {
	setting.InitConf(path)
	common.InitApp()
	//model.InitDB()
}
