package json

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/hb1707/ant-godmin/auth"
	"github.com/hb1707/ant-godmin/model"
	"github.com/hb1707/ant-godmin/pkg/log"
	"github.com/hb1707/ant-godmin/routers/hook"
	"github.com/hb1707/ant-godmin/service"
	"github.com/hb1707/exfun/fun"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"
)

var (
	ErrUploadFail = errors.New("接收文件失败")
)

func UploadFile(c *gin.Context) {
	var file model.Files
	var req model.Files
	pathStr := c.Param("path")
	typeId, _ := strconv.Atoi(c.DefaultQuery("type_id", "0"))
	photoId, _ := strconv.Atoi(c.DefaultPostForm("photo_id", "0"))
	req.TypeId = uint(typeId)
	uid, _ := auth.Identity(c)
	req.Uid = uint(uid)
	req.From = "T"
	req.PhotoId = uint(photoId)
	_, header, err := c.Request.FormFile("file")
	if err != nil {
		log.Error("接收文件失败!", err)
		jsonErr(c, http.StatusInternalServerError, ErrUploadFail)
		return
	}
	hookReq := hook.GenerateFileTag(c)
	req.Tag = hookReq.Tag
	ext := path.Ext(header.Filename)
	name := strings.TrimSuffix(header.Filename, ext)
	name = fun.MD5(name)
	req.Name = name + "_" + time.Now().Format("20060102150405") + ext
	err, file = service.NewFileService().UploadFile(header, pathStr, req) // 文件上传后拿到文件路径
	if err != nil {
		log.Error("修改数据库链接失败!", err)
		jsonErr(c, http.StatusInternalServerError, err)
		return
	}
	if file.Id > 0 {
		jsonResult(c, http.StatusOK, map[string]interface{}{
			"file": file,
		})
		return
	} else {
		jsonResult(c, http.StatusOK, map[string]interface{}{})
		return
	}
}
