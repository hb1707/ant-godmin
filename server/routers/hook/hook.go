package hook

import (
	"github.com/gin-gonic/gin"
	"github.com/hb1707/ant-godmin/model"
)

var GenerateFileTag func(c *gin.Context) model.Files
