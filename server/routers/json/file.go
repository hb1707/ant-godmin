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
	typeId, _ := strconv.Atoi(c.DefaultPostForm("type_id", "0"))
	photoId, _ := strconv.Atoi(c.DefaultPostForm("photo_id", "0")) //todo: 20220621之后准备废弃
	fileId, _ := strconv.Atoi(c.DefaultPostForm("file_id", "0"))
	fileName := c.DefaultPostForm("file_name", "")
	req.TypeId = uint(typeId)
	uid, _ := auth.Identity(c)
	req.Uid = uint(uid)
	req.From = "T"
	if photoId > 0 {
		req.FileId = uint(photoId)
	} else {
		req.FileId = uint(fileId)
	}
	_, header, err := c.Request.FormFile("file")
	if err != nil {
		log.Error("接收文件失败!", err)
		jsonErr(c, http.StatusInternalServerError, ErrUploadFail)
		return
	}
	hookReq := hook.GenerateFileTag(c)
	req.Tag = hookReq.Tag
	ext := path.Ext(header.Filename)
	if fileName == "" {
		name := strings.TrimSuffix(header.Filename, ext)
		name = fun.MD5(name)
		fileName = name + "_" + time.Now().Format("20060102150405")
	}
	req.Name = fileName + ext
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
